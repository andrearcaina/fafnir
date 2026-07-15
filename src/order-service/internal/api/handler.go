package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"fafnir/order-service/internal/db"
	"fafnir/order-service/internal/db/generated"
	basepb "fafnir/shared/pb/base"
	orderpb "fafnir/shared/pb/order"
	stockpb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/logger"
	natsC "fafnir/shared/pkg/nats"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type OrderHandler struct {
	db          *db.Database
	natsClient  *natsC.NatsClient
	stockClient stockpb.StockServiceClient
	logger      *logger.Logger
	orderpb.UnimplementedOrderServiceServer
}

var errInvalidOrderEvent = errors.New("invalid order event")

const (
	eventProcessingTimeout  = 10 * time.Second
	maxTradableSymbolLength = 10 // order and portfolio schemas currently use VARCHAR(10)
	publishAttempts         = 3
	publishRetryDelay       = 100 * time.Millisecond
)

func NewOrderHandler(db *db.Database, natsClient *natsC.NatsClient, stockClient stockpb.StockServiceClient, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{
		db:          db,
		natsClient:  natsClient,
		stockClient: stockClient,
		logger:      logger,
	}
}

func (h *OrderHandler) RegisterSubscribeHandlers() {
	_, err := h.natsClient.QueueSubscribe("orders.>", "order-service", "order-service-durable", h.handleOrderEvents)
	if err != nil {
		h.logger.Debug(context.Background(), "Failed to subscribe to orders.> subject", "error", err)
	}
}

func (h *OrderHandler) handleOrderEvents(msg *nats.Msg) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), eventProcessingTimeout)
	defer cancel()

	switch msg.Subject {
	case "orders.filled":
		err = h.handleOrderFilled(ctx, msg)
	case "orders.rejected":
		err = h.handleOrderRejected(ctx, msg)
	default:
		// ignore events we don't care about
		// we must ack them, otherwise they come back forever
		_ = msg.Ack()
		return
	}

	if err != nil {
		h.logger.Error(ctx, "Failed to process message", "subject", msg.Subject, "error", err)
		if errors.Is(err, errInvalidOrderEvent) {
			_ = msg.Term()
			return
		}

		_ = msg.NakWithDelay(2 * time.Second)
	} else {
		_ = msg.Ack() // success (acknowledge message)
	}
}

func (h *OrderHandler) GetOrderById(ctx context.Context, req *orderpb.GetOrderByIdRequest) (*orderpb.GetOrderByIdResponse, error) {
	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return &orderpb.GetOrderByIdResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &orderpb.GetOrderByIdResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("invalid user ID")
	}

	order, err := h.db.GetQueries().GetOrderByIdAndUserId(ctx, generated.GetOrderByIdAndUserIdParams{
		ID:     orderId,
		UserID: userID,
	})
	if err != nil {
		return &orderpb.GetOrderByIdResponse{
			Code: basepb.ErrorCode_NOT_FOUND,
		}, err
	}

	return &orderpb.GetOrderByIdResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}

func (h *OrderHandler) GetOrdersByUserId(ctx context.Context, req *orderpb.GetOrdersByUserIdRequest) (*orderpb.GetOrdersByUserIdResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &orderpb.GetOrdersByUserIdResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	orders, err := h.db.GetQueries().GetOrdersByUserId(ctx, userId)
	if err != nil {
		return &orderpb.GetOrdersByUserIdResponse{
			Code: basepb.ErrorCode_NOT_FOUND,
		}, err
	}

	var responseOrders []*orderpb.Order
	for _, order := range orders {
		responseOrders = append(responseOrders, convertOrderToProto(order))
	}

	return &orderpb.GetOrdersByUserIdResponse{
		Code:   basepb.ErrorCode_OK,
		Orders: responseOrders,
	}, nil
}

func (h *OrderHandler) InsertOrder(ctx context.Context, req *orderpb.InsertOrderRequest) (*orderpb.InsertOrderResponse, error) {
	if req.Side != orderpb.OrderSide_ORDER_SIDE_BUY && req.Side != orderpb.OrderSide_ORDER_SIDE_SELL {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("side must be BUY or SELL")
	}

	if req.Type != orderpb.OrderType_ORDER_TYPE_MARKET && req.Type != orderpb.OrderType_ORDER_TYPE_LIMIT {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("only MARKET and LIMIT orders are supported")
	}

	if !isPositiveFinite(req.Quantity) {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("quantity must be greater than zero")
	}
	if req.Type == orderpb.OrderType_ORDER_TYPE_LIMIT && !isPositiveFinite(req.Price) {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("limit price must be greater than zero")
	}
	if req.StopPrice != 0 {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("stop prices are not supported")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("invalid user ID")
	}

	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" || len(symbol) > maxTradableSymbolLength || strings.ContainsAny(symbol, " \t\r\n") {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("symbol is not supported for trading")
	}
	metadata, err := h.stockClient.GetStockMetadata(ctx, &stockpb.GetStockMetadataRequest{
		Symbol: symbol,
	})
	if err != nil {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, fmt.Errorf("validate symbol: %w", err)
	}
	if metadata.Code != basepb.ErrorCode_OK || metadata.Data == nil {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("invalid symbol")
	}
	if !isTradableInstrument(metadata.Data.InstrumentType) {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, fmt.Errorf("%s instruments are not supported for trading", metadata.Data.InstrumentType)
	}

	params := generated.InsertOrderParams{
		UserID:    userID,
		Symbol:    symbol,
		Side:      convertOrderSideToDB(req.Side),
		Type:      convertOrderTypeToDB(req.Type),
		Status:    generated.OrderStatusPending,
		Quantity:  floatToNumeric(req.Quantity),
		Price:     floatToNumericNullIfZero(req.Price),
		StopPrice: floatToNumericNullIfZero(req.StopPrice),
	}

	order, err := h.db.GetQueries().InsertOrder(ctx, params)
	if err != nil {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	// publish order created event
	event := &orderpb.OrderCreatedEvent{
		OrderId:   order.ID.String(),
		UserId:    order.UserID.String(),
		Symbol:    order.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Status:    orderpb.OrderStatus_ORDER_STATUS_PENDING,
		Quantity:  req.Quantity,
		Price:     req.Price,
		StopPrice: req.StopPrice,
		CreatedAt: convertTime(order.CreatedAt),
	}

	eventBytes, err := proto.Marshal(event)
	if err != nil {
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INTERNAL}, fmt.Errorf("marshal orders.created event: %w", err)
	}
	if err := h.publishEvent(ctx, "orders.created", order.ID.String()+":created", eventBytes); err != nil {
		_, rejectErr := h.db.GetQueries().RejectOrder(ctx, order.ID)
		return &orderpb.InsertOrderResponse{Code: basepb.ErrorCode_INTERNAL}, errors.Join(
			fmt.Errorf("publish orders.created event: %w", err),
			rejectErr,
		)
	}

	return &orderpb.InsertOrderResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}

func isTradableInstrument(instrumentType string) bool {
	switch strings.ToUpper(instrumentType) {
	case "EQUITY", "ETF":
		return true
	default:
		return false
	}
}

func isPositiveFinite(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return &orderpb.CancelOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &orderpb.CancelOrderResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, errors.New("invalid user ID")
	}

	order, err := h.db.GetQueries().CancelOrder(ctx, generated.CancelOrderParams{
		ID:     orderId,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &orderpb.CancelOrderResponse{
				Code: basepb.ErrorCode_NOT_FOUND,
			}, errors.New("order not found or not in pending status")
		}
		return &orderpb.CancelOrderResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	// publish order cancelled event
	event := &orderpb.OrderCancelledEvent{
		OrderId:     order.ID.String(),
		UserId:      order.UserID.String(),
		Symbol:      order.Symbol,
		Side:        convertOrderSide(order.Side),
		Status:      convertOrderStatus(order.Status),
		CancelledAt: convertTime(order.UpdatedAt),
	}

	eventBytes, err := proto.Marshal(event)
	if err != nil {
		return &orderpb.CancelOrderResponse{Code: basepb.ErrorCode_INTERNAL}, fmt.Errorf("marshal orders.cancelled event: %w", err)
	}
	if err := h.publishEvent(ctx, "orders.cancelled", order.ID.String()+":cancelled", eventBytes); err != nil {
		return &orderpb.CancelOrderResponse{Code: basepb.ErrorCode_INTERNAL}, fmt.Errorf("publish orders.cancelled event: %w", err)
	}

	return &orderpb.CancelOrderResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}

func (h *OrderHandler) publishEvent(ctx context.Context, subject string, messageID string, data []byte) error {
	var lastErr error
	for attempt := 1; attempt <= publishAttempts; attempt++ {
		if _, err := h.natsClient.PublishWithID(subject, messageID, data); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == publishAttempts {
			break
		}
		timer := time.NewTimer(publishRetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}

func (h *OrderHandler) handleOrderFilled(ctx context.Context, msg *nats.Msg) error {
	var event orderpb.OrderFilledEvent
	if err := proto.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Debug(ctx, "Error unmarshalling order filled event", "error", err)
		return fmt.Errorf("%w: decode filled event: %v", errInvalidOrderEvent, err)
	}

	orderId, err := uuid.Parse(event.OrderId)
	if err != nil {
		h.logger.Debug(ctx, "Invalid order ID in filled event", "error", err)
		return fmt.Errorf("%w: invalid filled order ID", errInvalidOrderEvent)
	}
	if !isPositiveFinite(event.FillQuantity) || !isPositiveFinite(event.FillPrice) {
		return fmt.Errorf("%w: fill quantity and price must be greater than zero", errInvalidOrderEvent)
	}

	filledAt := time.Now().UTC()
	if event.FilledAt != nil && event.FilledAt.IsValid() {
		filledAt = event.FilledAt.AsTime()
	}

	err = h.db.ExecMultiTx(ctx, func(queries *generated.Queries) error {
		order, err := queries.GetOrderByIdForUpdate(ctx, orderId)
		if err != nil {
			return err
		}
		if order.Status == generated.OrderStatusFilled {
			return nil
		}
		if order.Status != generated.OrderStatusPending {
			h.logger.Info(ctx, "Ignoring fill for terminal order", "order_id", event.OrderId, "status", order.Status)
			return nil
		}

		if err := queries.InsertOrderFilled(ctx, generated.InsertOrderFilledParams{
			OrderID:      orderId,
			FillQuantity: floatToNumeric(event.FillQuantity),
			FillPrice:    floatToNumeric(event.FillPrice),
			FilledAt:     pgtype.Timestamptz{Time: filledAt, Valid: true},
		}); err != nil {
			return fmt.Errorf("insert order fill: %w", err)
		}

		_, err = queries.UpdateOrderStatus(ctx, generated.UpdateOrderStatusParams{
			ID:             orderId,
			FilledQuantity: floatToNumeric(event.FillQuantity),
			AvgFillPrice:   floatToNumeric(event.FillPrice),
			Status:         generated.OrderStatusFilled,
		})

		return err
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Debug(ctx, "Order not found during fill update", "order_id", event.OrderId)
			return fmt.Errorf("%w: filled order not found", errInvalidOrderEvent)
		}
		h.logger.Debug(ctx, "Failed to update order status", "order_id", event.OrderId, "error", err)
		return err
	}

	h.logger.Info(ctx, "Order updated to FILLED", "order_id", event.OrderId)
	return nil
}

func (h *OrderHandler) handleOrderRejected(ctx context.Context, msg *nats.Msg) error {
	var event orderpb.OrderRejectedEvent
	if err := proto.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Debug(ctx, "Error unmarshalling order rejected event", "error", err)
		return fmt.Errorf("%w: decode rejected event: %v", errInvalidOrderEvent, err)
	}

	orderId, err := uuid.Parse(event.OrderId)
	if err != nil {
		h.logger.Debug(ctx, "Invalid order ID in rejected event", "error", err)
		return fmt.Errorf("%w: invalid rejected order ID", errInvalidOrderEvent)
	}

	_, err = h.db.GetQueries().RejectOrder(ctx, orderId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(ctx, "Ignoring rejection for terminal order", "order_id", event.OrderId)
			return nil
		}
		h.logger.Debug(ctx, "Failed to update order status to rejected", "order_id", event.OrderId, "error", err)
		return err
	}

	h.logger.Info(ctx, "Order updated to REJECTED", "order_id", event.OrderId, "reason", event.Reason)
	return nil
}
