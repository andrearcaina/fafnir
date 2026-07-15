package engine

import (
	"fmt"
	"math"
	"strings"

	orderpb "fafnir/shared/pb/order"
	portfoliopb "fafnir/shared/pb/portfolio"
)

func isTradableInstrument(instrumentType string) bool {
	switch strings.ToUpper(instrumentType) {
	case "EQUITY", "ETF":
		return true
	default:
		return false
	}
}

type decision uint8

const (
	decisionWait decision = iota
	decisionFill
)

func validateOrder(order *orderpb.OrderCreatedEvent) error {
	if order.OrderId == "" || order.UserId == "" || order.Symbol == "" {
		return fmt.Errorf("order identifiers and symbol are required")
	}
	if !positiveFinite(order.Quantity) {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if order.Side != orderpb.OrderSide_ORDER_SIDE_BUY && order.Side != orderpb.OrderSide_ORDER_SIDE_SELL {
		return fmt.Errorf("unsupported order side")
	}

	switch order.Type {
	case orderpb.OrderType_ORDER_TYPE_MARKET:
		return nil
	case orderpb.OrderType_ORDER_TYPE_LIMIT:
		if !positiveFinite(order.Price) {
			return fmt.Errorf("limit price must be greater than zero")
		}
		return nil
	default:
		return fmt.Errorf("only market and limit orders are supported")
	}
}

func positiveFinite(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func evaluateOrder(order *orderpb.OrderCreatedEvent, currentPrice float64) decision {
	if order.Type == orderpb.OrderType_ORDER_TYPE_MARKET {
		return decisionFill
	}

	if order.Side == orderpb.OrderSide_ORDER_SIDE_BUY && currentPrice <= order.Price {
		return decisionFill
	}
	if order.Side == orderpb.OrderSide_ORDER_SIDE_SELL && currentPrice >= order.Price {
		return decisionFill
	}

	return decisionWait
}

func currencyCode(currency portfoliopb.CurrencyType) (string, error) {
	switch currency {
	case portfoliopb.CurrencyType_CURRENCY_TYPE_USD:
		return "USD", nil
	case portfoliopb.CurrencyType_CURRENCY_TYPE_CAD:
		return "CAD", nil
	default:
		return "", fmt.Errorf("account currency is unsupported")
	}
}
