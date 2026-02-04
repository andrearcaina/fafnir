package api

import (
	"context"
	"errors"
	"fmt"

	"fafnir/order-service/internal/db"
	"fafnir/order-service/internal/db/generated"
	basepb "fafnir/shared/pb/base"
	orderpb "fafnir/shared/pb/order"
	stockpb "fafnir/shared/pb/stock"
	natsC "fafnir/shared/pkg/nats"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type OrderHandler struct {
	db          *db.Database
	natsClient  *natsC.NatsClient
	stockClient stockpb.StockServiceClient
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(db *db.Database, natsClient *natsC.NatsClient, stockClient stockpb.StockServiceClient) *OrderHandler {
	return &OrderHandler{
		db:          db,
		natsClient:  natsClient,
		stockClient: stockClient,
	}
}

func (h *OrderHandler) ConsumeFilledEvents() {
	_, err := h.natsClient.QueueSubscribe("orders.filled", "order-service", "order-service-durable", func(msg *nats.Msg) {
		var event orderpb.OrderFilledEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			fmt.Printf("Error unmarshalling order filled event: %v\n", err)
			return
		}

		orderId, err := uuid.Parse(event.OrderId)
		if err != nil {
			fmt.Printf("Invalid order ID in filled event: %v\n", err)
			return
		}

		params := generated.UpdateOrderStatusParams{
			ID:             orderId,
			FilledQuantity: floatToNumeric(event.FillQuantity),
			AvgFillPrice:   floatToNumeric(event.FillPrice),
			Status:         generated.OrderStatusFilled,
		}

		_, err = h.db.GetQueries().UpdateOrderStatus(context.Background(), params)
		if err != nil {
			fmt.Printf("Failed to update order status for order %s: %v\n", event.OrderId, err)
			return
		}

		fmt.Printf("Order %s updated to FILLED via NATS event\n", event.OrderId)
	})

	if err != nil {
		fmt.Printf("Failed to subscribe to orders.filled: %v\n", err)
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
		Side:      generated.OrderSide(req.Side),
		Type:      generated.OrderType(req.Type),
		Status:    generated.OrderStatus(status),
		Quantity:  floatToNumeric(req.Quantity),
		Price:     floatToNumeric(req.Price),
		StopPrice: floatToNumeric(req.StopPrice),
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
		fmt.Printf("failed to marshal orders.created event: %v\n", err)
	} else {
		_, err = h.natsClient.Publish("orders.created", eventBytes)
		if err != nil {
			fmt.Printf("failed to publish orders.created event: %v\n", err)
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
			}, fmt.Errorf("order not found or not in pending status")
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
		fmt.Printf("failed to marshal orders.cancelled event: %v\n", err)
	} else {
		_, err = h.natsClient.Publish("orders.cancelled", eventBytes)
		if err != nil {
			fmt.Printf("failed to publish orders.cancelled event: %v\n", err)
		}
	}

	return &orderpb.CancelOrderResponse{
		Code:  basepb.ErrorCode_OK,
		Order: convertOrderToProto(order),
	}, nil
}
