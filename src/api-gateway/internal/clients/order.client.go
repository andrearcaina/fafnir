package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	"strings"

	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	conn   *grpc.ClientConn
	client pb.OrderServiceClient
}

func NewOrderClient(address string) *OrderClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil
	}

	client := pb.NewOrderServiceClient(conn)

	return &OrderClient{
		conn:   conn,
		client: client,
	}
}

func (c *OrderClient) InsertOrder(ctx context.Context, userID string, input model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	// Defaulting enums or values if they are simple strings in GraphQL
	sideKey := "ORDER_SIDE_" + input.Side
	sideVal, ok := pb.OrderSide_value[sideKey]
	if !ok {
		// If not found, try uppercase
		sideKey = "ORDER_SIDE_" + strings.ToUpper(input.Side)
		sideVal, ok = pb.OrderSide_value[sideKey]
		if !ok {
			return model.CreateOrderResponse{
				Code: basepb.ErrorCode_INVALID_ARGUMENT.String(),
			}, nil // Return error or handle gracefully
		}
	}
	side := pb.OrderSide(sideVal)

	typeKey := "ORDER_TYPE_" + input.Type
	typeVal, ok := pb.OrderType_value[typeKey]
	if !ok {
		// If not found, try uppercase
		typeKey = "ORDER_TYPE_" + strings.ToUpper(input.Type)
		typeVal, ok = pb.OrderType_value[typeKey]
		if !ok {
			return model.CreateOrderResponse{
				Code: basepb.ErrorCode_INVALID_ARGUMENT.String(),
			}, nil
		}
	}
	type_ := pb.OrderType(typeVal)

	req := &pb.InsertOrderRequest{
		UserId:    userID,
		Symbol:    input.Symbol,
		Side:      side,
		Type:      type_,
		Quantity:  input.Quantity,
		Price:     safeFloat(input.Price),
		StopPrice: safeFloat(input.StopPrice),
	}

	resp, err := c.client.InsertOrder(ctx, req)
	if err != nil {
		return model.CreateOrderResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.CreateOrderResponse{
		Data: mapProtoToModel(resp.Order),
		Code: resp.GetCode().String(),
	}, nil
}

func (c *OrderClient) CancelOrder(ctx context.Context, orderID, userID string) (model.CancelOrderResponse, error) {
	req := &pb.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	}

	resp, err := c.client.CancelOrder(ctx, req)
	if err != nil {
		return model.CancelOrderResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.CancelOrderResponse{
		Data: mapProtoToModel(resp.Order),
		Code: resp.GetCode().String(),
	}, nil
}

func (c *OrderClient) GetOrders(ctx context.Context, userID string) (model.OrdersResponse, error) {
	req := &pb.GetOrdersByUserIdRequest{
		UserId: userID,
	}

	resp, err := c.client.GetOrdersByUserId(ctx, req)
	if err != nil {
		return model.OrdersResponse{
			Code:  basepb.ErrorCode_INTERNAL.String(),
			Count: 0,
		}, err
	}

	var orders []*model.Order
	for _, order := range resp.Orders {
		orders = append(orders, mapProtoToModel(order))
	}

	if len(orders) == 0 {
		return model.OrdersResponse{
			Code:  basepb.ErrorCode_NOT_FOUND.String(),
			Count: 0,
		}, nil
	}

	return model.OrdersResponse{
		Data:  orders,
		Count: int32(len(orders)),
		Code:  resp.GetCode().String(),
	}, nil
}

// Helpers

func safeFloat(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func mapProtoToModel(o *pb.Order) *model.Order {
	if o == nil {
		return nil
	}
	return &model.Order{
		ID:             o.Id,
		UserID:         o.UserId,
		Symbol:         o.Symbol,
		Side:           o.Side.String(),
		Type:           o.Type.String(),
		Status:         o.Status.String(),
		Quantity:       o.Quantity,
		Price:          o.Price,
		StopPrice:      o.StopPrice,
		FilledQuantity: o.FilledQuantity,
		AvgFillPrice:   o.AvgFillPrice,
		CreatedAt:      o.CreatedAt.AsTime().String(),
		UpdatedAt:      o.UpdatedAt.AsTime().String(),
	}
}
