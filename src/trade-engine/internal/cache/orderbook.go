package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	orderpb "fafnir/shared/pb/order"
	"fafnir/shared/pkg/redis"
)

const (
	activeSymbolsKey = "orderbook:v2:active_symbols"
	addOrderScript   = `
redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
redis.call("SADD", KEYS[2], ARGV[3])
return 1
`
	removeOrderScript = `
local removed = redis.call("HDEL", KEYS[1], ARGV[1])
if redis.call("HLEN", KEYS[1]) == 0 then
    redis.call("SREM", KEYS[2], ARGV[2])
end
return removed
`
)

type OrderBook struct {
	client *redis.Cache
}

func NewOrderBook(client *redis.Cache) *OrderBook {
	return &OrderBook{client: client}
}

func (o *OrderBook) Add(ctx context.Context, order *orderpb.OrderCreatedEvent) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order %s: %w", order.OrderId, err)
	}

	key := ordersKey(order.Symbol)
	if _, err := o.client.Eval(
		ctx,
		addOrderScript,
		[]string{key, activeSymbolsKey},
		order.OrderId,
		string(data),
		order.Symbol,
	); err != nil {
		return fmt.Errorf("store order %s: %w", order.OrderId, err)
	}

	return nil
}

func (o *OrderBook) Symbols(ctx context.Context) ([]string, error) {
	symbols, err := o.client.SMembers(ctx, activeSymbolsKey)
	if err != nil {
		return nil, fmt.Errorf("list active symbols: %w", err)
	}

	return symbols, nil
}

func (o *OrderBook) ClaimMatched(ctx context.Context, symbol string, currentPrice float64) ([]*orderpb.OrderCreatedEvent, error) {
	key := ordersKey(symbol)
	rawOrders, err := o.client.HGetAll(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("list orders for %s: %w", symbol, err)
	}

	matched := make([]*orderpb.OrderCreatedEvent, 0)
	for orderID, rawOrder := range rawOrders {
		var order orderpb.OrderCreatedEvent
		if err := json.Unmarshal([]byte(rawOrder), &order); err != nil {
			_, removeErr := o.remove(ctx, symbol, orderID)
			return nil, errors.Join(
				fmt.Errorf("unmarshal order %s: %w", orderID, err),
				removeErr,
			)
		}

		if !matchesLimit(&order, currentPrice) {
			continue
		}

		removed, err := o.remove(ctx, symbol, orderID)
		if err != nil {
			return nil, fmt.Errorf("claim order %s: %w", orderID, err)
		}
		if removed == 1 {
			matched = append(matched, &order)
		}
	}

	return matched, nil
}

func (o *OrderBook) Remove(ctx context.Context, symbol string, orderID string) error {
	if _, err := o.remove(ctx, symbol, orderID); err != nil {
		return fmt.Errorf("remove order %s: %w", orderID, err)
	}

	return nil
}

func (o *OrderBook) remove(ctx context.Context, symbol string, orderID string) (int64, error) {
	result, err := o.client.Eval(
		ctx,
		removeOrderScript,
		[]string{ordersKey(symbol), activeSymbolsKey},
		orderID,
		symbol,
	)
	if err != nil {
		return 0, err
	}

	removed, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected Redis result %T", result)
	}

	return removed, nil
}

func ordersKey(symbol string) string {
	return fmt.Sprintf("orderbook:v2:orders:%s", symbol)
}

func matchesLimit(order *orderpb.OrderCreatedEvent, currentPrice float64) bool {
	switch order.Side {
	case orderpb.OrderSide_ORDER_SIDE_BUY:
		return currentPrice <= order.Price
	case orderpb.OrderSide_ORDER_SIDE_SELL:
		return currentPrice >= order.Price
	default:
		return false
	}
}
