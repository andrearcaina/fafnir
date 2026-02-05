package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	orderpb "fafnir/shared/pb/order"
	stockpb "fafnir/shared/pb/stock"
	natsC "fafnir/shared/pkg/nats"
	"fafnir/trade-engine/internal/config"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Engine struct {
	cfg         *config.Config
	natsClient  *natsC.NatsClient
	stockClient stockpb.StockServiceClient
	orderBook   *OrderBook // basically a "cache" of pending limit orders
	stopCh      chan struct{}
}

func NewEngine(cfg *config.Config) (*Engine, error) {
	log.Printf("Engine connecting to NATS at %s", cfg.NATS.URL)
	nc, err := natsC.New(cfg.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	log.Printf("Engine connecting to Stock Service at %s", cfg.StockService.URL)
	conn, err := grpc.NewClient(cfg.StockService.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stock service: %w", err)
	}
	stockClient := stockpb.NewStockServiceClient(conn)

	return &Engine{
		cfg:         cfg,
		natsClient:  nc,
		stockClient: stockClient,
		orderBook:   NewOrderBook(),
		stopCh:      make(chan struct{}),
	}, nil
}

func (e *Engine) Start() {
	err := e.subscribeToOrders()
	if err != nil {
		log.Fatalf("Failed to subscribe to orders.created: %v", err)
	}

	// start polling for pending limit orders
	go e.pollOrders()

	// basically block (the main thread) from exiting until stopCh is closed
	<-e.stopCh
}

func (e *Engine) Stop() error {
	close(e.stopCh)
	if e.natsClient != nil {
		e.natsClient.Close()
	}

	return nil
}

func (e *Engine) subscribeToOrders() error {
	_, err := e.natsClient.QueueSubscribe("orders.created", "trade-engine", "trade-engine-durable", func(msg *nats.Msg) {
		var event orderpb.OrderCreatedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshalling order event: %v", err)
			return
		}

		log.Printf("Received Order: ID=%s Symbol=%s Type=%v", event.OrderId, event.Symbol, event.Type)
		e.processOrder(&event)
	})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (e *Engine) pollOrders() {
	ticker := time.NewTicker(5 * time.Second) // poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			symbols := e.orderBook.MGet()
			if len(symbols) == 0 {
				continue
			}

			// batch process all symbols
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			resp, err := e.stockClient.GetStockQuoteBatch(ctx, &stockpb.GetStockQuoteBatchRequest{
				Symbols: symbols,
			})
			cancel()

			if err != nil {
				log.Printf("Failed to get stock quotes batch: %v", err)
				continue
			}

			// process results for each symbol
			for _, quote := range resp.Data {
				filledOrders := e.orderBook.Evaluate(quote.Symbol, quote.LastPrice)
				for _, order := range filledOrders {
					e.publishFilledEvent(order, quote.LastPrice)
				}
			}
		}
	}
}

func (e *Engine) processOrder(order *orderpb.OrderCreatedEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get current market price
	resp, err := e.stockClient.GetStockQuote(ctx, &stockpb.GetStockQuoteRequest{
		Symbol: order.Symbol,
	})
	if err != nil {
		log.Printf("Failed to get stock quote for %s: %v", order.Symbol, err)
		return
	}

	currentPrice := resp.Data.LastPrice
	shouldFill := false

	// evaluate match
	switch order.Type {
	case orderpb.OrderType_ORDER_TYPE_MARKET:
		shouldFill = true
	case orderpb.OrderType_ORDER_TYPE_LIMIT:
		switch order.Side {
		case orderpb.OrderSide_ORDER_SIDE_BUY:
			// buy limit: valid if current price is <= limit price
			if currentPrice <= order.Price {
				shouldFill = true
			}
		case orderpb.OrderSide_ORDER_SIDE_SELL:
			// sell limit: valid if current price is >= limit price
			if currentPrice >= order.Price {
				shouldFill = true
			}
		}
	}

	if shouldFill {
		e.publishFilledEvent(order, currentPrice)
	} else {
		// if not filled, add to order book
		if order.Type == orderpb.OrderType_ORDER_TYPE_LIMIT {
			log.Printf("Order %s not immediately filled. Adding to Order Book. Limit: %f, Current: %f",
				order.OrderId, order.Price, currentPrice)
			e.orderBook.Add(order)
		} else {
			// log market order not filled (unexpected)
			// this should never happen and honestly, it's not a big deal (as it's a simulator)
			log.Printf("Market Order %s not filled (unexpected).", order.OrderId)
		}
	}
}

func (e *Engine) publishFilledEvent(order *orderpb.OrderCreatedEvent, fillPrice float64) {
	filledEvent := &orderpb.OrderFilledEvent{
		OrderId:      order.OrderId,
		UserId:       order.UserId,
		Symbol:       order.Symbol,
		FillQuantity: order.Quantity,
		FillPrice:    fillPrice,
		FilledAt:     timestamppb.Now(),
	}

	data, err := proto.Marshal(filledEvent)
	if err != nil {
		log.Printf("Failed to marshal filled event: %v", err)
		return
	}

	_, err = e.natsClient.Publish("orders.filled", data)
	if err != nil {
		log.Printf("Failed to publish orders.filled for order %s: %v", order.OrderId, err)
		return
	}

	log.Printf("Order %s FILLED at %f", order.OrderId, fillPrice)
}
