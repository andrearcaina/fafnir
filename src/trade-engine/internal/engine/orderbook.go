package engine

import (
	"sync"

	orderpb "fafnir/shared/pb/order"
)

// technically, this is a "map" (cache) of "lists" (queue of orders)
type OrderBook struct {
	mu     sync.Mutex
	orders map[string][]*orderpb.OrderCreatedEvent // symbol -> list of orders
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		orders: make(map[string][]*orderpb.OrderCreatedEvent),
	}
}

func (o *OrderBook) Add(order *orderpb.OrderCreatedEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.orders[order.Symbol] == nil {
		o.orders[order.Symbol] = make([]*orderpb.OrderCreatedEvent, 0)
	}
	o.orders[order.Symbol] = append(o.orders[order.Symbol], order)
}

func (o *OrderBook) MGet() []string {
	o.mu.Lock()
	defer o.mu.Unlock()

	symbols := make([]string, 0, len(o.orders))
	for s := range o.orders {
		symbols = append(symbols, s)
	}
	return symbols
}

func (o *OrderBook) Evaluate(symbol string, currentPrice float64) []*orderpb.OrderCreatedEvent {
	o.mu.Lock()
	defer o.mu.Unlock()

	orders, exists := o.orders[symbol]
	if !exists || len(orders) == 0 {
		return nil
	}

	var filled []*orderpb.OrderCreatedEvent
	var remaining []*orderpb.OrderCreatedEvent

	// can optimize later (e.g. binary search)
	for _, order := range orders {
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
			filled = append(filled, order)
		} else {
			remaining = append(remaining, order)
		}
	}

	// update the book with remaining orders
	if len(remaining) == 0 {
		delete(o.orders, symbol)
	} else {
		o.orders[symbol] = remaining
	}

	return filled
}
