package api

import (
	"context"
	"errors"
	"fmt"
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

	switch msg.Subject {
	case "orders.filled":
		err = h.handleOrderFilled(msg)
	case "orders.rejected":
		err = h.handleOrderRejected(msg)
	default:
		// ignore events we don't care about
		// we must ack them, otherwise they come back forever
		_ = msg.Ack()
		return
	}

	if err != nil {
		h.logger.Error(context.Background(), "Failed to process message", "subject", msg.Subject, "error", err)
		_ = msg.Nak() // retry later (negative ack)
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

	order, err := h.db.GetQueries().GetOrderById(ctx, orderId)
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
	// validation, make sure side and type are not unspecified
	if req.Side == orderpb.OrderSide_ORDER_SIDE_UNSPECIFIED {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("side must be specified")
	}

	if req.Type == orderpb.OrderType_ORDER_TYPE_UNSPECIFIED {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("type must be specified")
	}

	// validate symbol with stock service
	_, err := h.stockClient.GetStockMetadata(ctx, &stockpb.GetStockMetadataRequest{
		Symbol: req.Symbol,
	})
	if err != nil {
		return &orderpb.InsertOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, fmt.Errorf("invalid symbol: %v", err)
	}

	status := req.Status
	if status == orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		status = orderpb.OrderStatus_ORDER_STATUS_PENDING
	}

	params := generated.InsertOrderParams{
		UserID:    uuid.MustParse(req.UserId),
		Symbol:    req.Symbol,
		Side:      convertOrderSideToDB(req.Side),
		Type:      convertOrderTypeToDB(req.Type),
		Status:    convertOrderStatusToDB(status),
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
		Quantity:  req.Quantity,
		Price:     req.Price,
		CreatedAt: convertTime(order.CreatedAt),
	}

	eventBytes, err := proto.Marshal(event)
	if err != nil {
		h.logger.Error(context.Background(), "Failed to marshal orders.created event", "error", err)
	} else {
		_, err = h.natsClient.Publish("orders.created", eventBytes)
		if err != nil {
			h.logger.Error(context.Background(), "Failed to publish orders.created event", "error", err)
		}
	}

	return &orderpb.InsertOrderResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return &orderpb.CancelOrderResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	order, err := h.db.GetQueries().CancelOrder(ctx, orderId)
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
		h.logger.Debug(context.Background(), "Failed to marshal orders.cancelled event", "error", err)
	} else {
		_, err = h.natsClient.Publish("orders.cancelled", eventBytes)
		if err != nil {
			h.logger.Debug(context.Background(), "Failed to publish orders.cancelled event", "error", err)
		}
	}

	return &orderpb.CancelOrderResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}

func (h *OrderHandler) handleOrderFilled(msg *nats.Msg) error {
	var event orderpb.OrderFilledEvent
	if err := proto.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Debug(context.Background(), "Error unmarshalling order filled event", "error", err)
		return err
	}

	orderId, err := uuid.Parse(event.OrderId)
	if err != nil {
		h.logger.Debug(context.Background(), "Invalid order ID in filled event", "error", err)
		return err
	}

	// insert into orders_fill table
	fillParams := generated.InsertOrderFilledParams{
		OrderID:      orderId,
		FillQuantity: floatToNumeric(event.FillQuantity),
		FillPrice:    floatToNumeric(event.FillPrice),
		FilledAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	if _, err := h.db.GetQueries().InsertOrderFilled(context.Background(), fillParams); err != nil {
		h.logger.Debug(context.Background(), "Failed to insert order fill", "error", err)
		// for now just log the error and continue
	}

	// update parent order status
	params := generated.UpdateOrderStatusParams{
		ID:             orderId,
		FilledQuantity: floatToNumeric(event.FillQuantity),
		AvgFillPrice:   floatToNumeric(event.FillPrice),
		Status:         generated.OrderStatusFilled,
	}

	_, err = h.db.GetQueries().UpdateOrderStatus(context.Background(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Debug(context.Background(), "Order not found during fill update", "order_id", event.OrderId)
			return err
		}
		h.logger.Debug(context.Background(), "Failed to update order status", "order_id", event.OrderId, "error", err)
		return err
	}

	h.logger.Info(context.Background(), "Order updated to FILLED", "order_id", event.OrderId)
	return nil
}

func (h *OrderHandler) handleOrderRejected(msg *nats.Msg) error {
	var event orderpb.OrderRejectedEvent
	if err := proto.Unmarshal(msg.Data, &event); err != nil {
		h.logger.Debug(context.Background(), "Error unmarshalling order rejected event", "error", err)
		return err
	}

	orderId, err := uuid.Parse(event.OrderId)
	if err != nil {
		h.logger.Debug(context.Background(), "Invalid order ID in rejected event", "error", err)
		return err
	}

	_, err = h.db.GetQueries().RejectOrder(context.Background(), orderId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Debug(context.Background(), "Order not found during rejection update", "order_id", event.OrderId)
			return err
		}
		h.logger.Debug(context.Background(), "Failed to update order status to rejected", "order_id", event.OrderId, "error", err)
		return err
	}

	h.logger.Info(context.Background(), "Order updated to REJECTED", "order_id", event.OrderId, "reason", event.Reason)
	return nil
}
