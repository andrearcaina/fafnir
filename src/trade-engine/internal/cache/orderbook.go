package cache

import (
	"context"
	"encoding/json"
	"fafnir/shared/pkg/redis"
	"fmt"
	"log"

	orderpb "fafnir/shared/pb/order"
)

type OrderBook struct {
	client *redis.Cache
}

func NewOrderBook(client *redis.Cache) *OrderBook {
	return &OrderBook{
		client: client,
	}
}

func (o *OrderBook) Add(order *orderpb.OrderCreatedEvent) {
	ctx := context.Background()
	// serializing order to JSON (simple, human readable)
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("Failed to marshal order %s: %v", order.OrderId, err)
		return
	}

	// add to symbol's order list
	// add symbol to active symbols set
	key := fmt.Sprintf("orders:%s", order.Symbol)
	if err := o.client.RPush(ctx, key, string(data)); err != nil {
		log.Printf("Failed to push order to redis: %v", err)
		return
	}

	if err := o.client.SAdd(ctx, "active_symbols", order.Symbol); err != nil {
		log.Printf("Failed to add symbol to active set: %v", err)
	}
}

func (o *OrderBook) MGet() []string {
	ctx := context.Background()
	symbols, err := o.client.SMembers(ctx, "active_symbols")

	if err != nil {
		log.Printf("Failed to get active symbols: %v", err)
		return []string{}
	}

	return symbols
}

func (o *OrderBook) Evaluate(symbol string, currentPrice float64) []*orderpb.OrderCreatedEvent {
	ctx := context.Background()
	key := fmt.Sprintf("orders:%s", symbol)

	// fetch all orders
	// not the most efficient, but it works for now
	rawOrders, err := o.client.LRange(ctx, key, 0, -1)
	if err != nil {
		log.Printf("Failed to get orders for %s: %v", symbol, err)
		return nil
	}

	if len(rawOrders) == 0 {
		// remove from active set if empty
		o.client.SRem(ctx, "active_symbols", symbol)
		return nil
	}

	var filled []*orderpb.OrderCreatedEvent
	var remaining []*orderpb.OrderCreatedEvent

	for _, raw := range rawOrders {
		var order orderpb.OrderCreatedEvent
		if err := json.Unmarshal([]byte(raw.(string)), &order); err != nil {
			log.Printf("Failed to unmarshal order: %v", err)
			continue
		}

		shouldFill := false
		switch order.Side {
		case orderpb.OrderSide_ORDER_SIDE_BUY:
			if currentPrice <= order.Price {
				shouldFill = true
			}
		case orderpb.OrderSide_ORDER_SIDE_SELL:
			if currentPrice >= order.Price {
				shouldFill = true
			}
		}

		if shouldFill {
			filled = append(filled, &order)
		} else {
			remaining = append(remaining, &order)
		}
	}

	// rewrite the list if any matched
	if len(filled) > 0 {
		// clear list
		o.client.Del(ctx, key)

		// push back remaining
		if len(remaining) > 0 {
			for _, order := range remaining {
				data, _ := json.Marshal(order)
				o.client.RPush(ctx, key, string(data))
			}
		} else {
			// no remaining orders, remove from active set
			o.client.SRem(ctx, "active_symbols", symbol)
		}
	}

	return filled
}
