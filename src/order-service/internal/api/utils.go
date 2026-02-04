package api

import (
	"fafnir/order-service/internal/db/generated"
	pb "fafnir/shared/pb/order"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertNumeric(n pgtype.Numeric) float64 {
	val, _ := n.Float64Value()
	return val.Float64
}

func convertOrderToProto(order generated.Order) *pb.Order {
	return &pb.Order{
		Id:             order.ID.String(),
		UserId:         order.UserID.String(),
		Symbol:         order.Symbol,
		Side:           convertOrderSide(order.Side),
		Type:           convertOrderType(order.Type),
		Status:         convertOrderStatus(order.Status),
		Quantity:       convertNumeric(order.Quantity),
		FilledQuantity: convertNumeric(order.FilledQuantity),
		Price:          convertNumeric(order.Price),
		StopPrice:      convertNumeric(order.StopPrice),
		AvgFillPrice:   convertNumeric(order.AvgFillPrice),
		CreatedAt:      convertTime(order.CreatedAt),
		UpdatedAt:      convertTime(order.UpdatedAt),
	}
}

func convertTime(t pgtype.Timestamptz) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}
	return timestamppb.New(t.Time)
}

func convertOrderSide(s generated.OrderSide) pb.OrderSide {
	switch s {
	case generated.OrderSideBuy:
		return pb.OrderSide_ORDER_SIDE_BUY
	case generated.OrderSideSell:
		return pb.OrderSide_ORDER_SIDE_SELL
	default:
		return pb.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func convertOrderType(t generated.OrderType) pb.OrderType {
	switch t {
	case generated.OrderTypeMarket:
		return pb.OrderType_ORDER_TYPE_MARKET
	case generated.OrderTypeLimit:
		return pb.OrderType_ORDER_TYPE_LIMIT
	case generated.OrderTypeStop:
		return pb.OrderType_ORDER_TYPE_STOP
	case generated.OrderTypeStopLimit:
		return pb.OrderType_ORDER_TYPE_STOP_LIMIT
	default:
		return pb.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

func convertOrderStatus(s generated.OrderStatus) pb.OrderStatus {
	switch s {
	case generated.OrderStatusPending:
		return pb.OrderStatus_ORDER_STATUS_PENDING
	case generated.OrderStatusPartiallyFilled:
		return pb.OrderStatus_ORDER_STATUS_PARTIAL_FILL
	case generated.OrderStatusFilled:
		return pb.OrderStatus_ORDER_STATUS_FILLED
	case generated.OrderStatusCanceled:
		return pb.OrderStatus_ORDER_STATUS_CANCELED
	case generated.OrderStatusRejected:
		return pb.OrderStatus_ORDER_STATUS_REJECTED
	default:
		return pb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(f); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}
