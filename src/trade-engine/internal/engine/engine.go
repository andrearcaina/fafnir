package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	orderpb "fafnir/shared/pb/order"
	portfoliopb "fafnir/shared/pb/portfolio"
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
	cfg             *config.Config
	natsClient      *natsC.NatsClient
	stockClient     stockpb.StockServiceClient
	portfolioClient portfoliopb.PortfolioServiceClient
	orderBook       *OrderBook // basically a "cache" of pending limit orders
	stopCh          chan struct{}
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

	log.Printf("Engine connecting to Portfolio Service at %s", cfg.Portfolio.URL)
	pConn, err := grpc.NewClient(cfg.Portfolio.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to portfolio service: %w", err)
	}
	portfolioClient := portfoliopb.NewPortfolioServiceClient(pConn)

	return &Engine{
		cfg:             cfg,
		natsClient:      nc,
		stockClient:     stockClient,
		portfolioClient: portfolioClient,
		orderBook:       NewOrderBook(),
		stopCh:          make(chan struct{}),
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
		_ = msg.Ack()
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
					hasSufficientResources := false
					if order.Side == orderpb.OrderSide_ORDER_SIDE_BUY {
						// buy
						requiredAmount := quote.LastPrice * order.Quantity
						if e.checkFunds(ctx, order.UserId, requiredAmount) {
							hasSufficientResources = true
						} else {
							log.Printf("Limit Order %s matched but insufficient funds (Required: %f). Rejecting.", order.OrderId, requiredAmount)
							e.publishRejectedEvent(order, "Insufficient funds")
						}
					} else {
						// sell
						if e.checkHoldings(ctx, order.UserId, order.Symbol, order.Quantity) {
							hasSufficientResources = true
						} else {
							log.Printf("Limit Order %s matched but insufficient holdings. Rejecting.", order.OrderId)
							e.publishRejectedEvent(order, "Insufficient holdings")
						}
					}

					if hasSufficientResources {
						e.publishFilledEvent(order, quote.LastPrice)
					} else {
						// put it back in the book so it can be retried later
						e.orderBook.Add(order)
					}
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

	// check funds/holdings before evaluating match
	hasSufficientResources := false
	if order.Side == orderpb.OrderSide_ORDER_SIDE_BUY {
		requiredAmount := currentPrice * order.Quantity
		if e.checkFunds(ctx, order.UserId, requiredAmount) {
			hasSufficientResources = true
		} else {
			log.Printf("Order %s rejected: Insufficient funds (Required: %f)", order.OrderId, requiredAmount)
			e.publishRejectedEvent(order, fmt.Sprintf("Insufficient funds. Required: %f", requiredAmount))
			return
		}
	} else { // order side is sell
		if e.checkHoldings(ctx, order.UserId, order.Symbol, order.Quantity) {
			hasSufficientResources = true
		} else {
			log.Printf("Order %s rejected: Insufficient holdings", order.OrderId)
			e.publishRejectedEvent(order, "Insufficient holdings")
			return
		}
	}

	if !hasSufficientResources {
		return
	}

	// market check: does the order price match the current market price?
	shouldExecute := false
	switch order.Type {
	case orderpb.OrderType_ORDER_TYPE_MARKET:
		shouldExecute = true
	case orderpb.OrderType_ORDER_TYPE_LIMIT:
		switch order.Side {
		case orderpb.OrderSide_ORDER_SIDE_BUY:
			// buy limit: execute if current price is cheaper or equal to limit
			if currentPrice <= order.Price {
				shouldExecute = true
			}
		case orderpb.OrderSide_ORDER_SIDE_SELL:
			// sell limit: execute if current price is higher or equal to limit
			if currentPrice >= order.Price {
				shouldExecute = true
			}
		}
	}

	// execute or queue (in order book)
	if shouldExecute {
		e.publishFilledEvent(order, currentPrice)
	} else {
		// no match yet (limit order waiting for price target)
		if order.Type == orderpb.OrderType_ORDER_TYPE_LIMIT {
			log.Printf("Order %s not immediately filled. Adding to Order Book. Limit: %f, Current: %f",
				order.OrderId, order.Price, currentPrice)
			e.orderBook.Add(order)
		} else {
			log.Printf("Market Order %s not filled (unexpected).", order.OrderId)
		}
	}
}

func (e *Engine) publishRejectedEvent(order *orderpb.OrderCreatedEvent, reason string) {
	rejectedEvent := &orderpb.OrderRejectedEvent{
		OrderId:    order.OrderId,
		UserId:     order.UserId,
		Symbol:     order.Symbol,
		Reason:     reason,
		RejectedAt: timestamppb.Now(),
	}

	data, err := proto.Marshal(rejectedEvent)
	if err != nil {
		log.Printf("Failed to marshal rejected event: %v", err)
		return
	}

	_, err = e.natsClient.Publish("orders.rejected", data)
	if err != nil {
		log.Printf("Failed to publish orders.rejected for order %s: %v", order.OrderId, err)
		return
	}

	log.Printf("Order %s REJECTED: %s", order.OrderId, reason)
}

func (e *Engine) checkFunds(ctx context.Context, userId string, requiredAmount float64) bool {
	resp, err := e.portfolioClient.GetPortfolioSummary(ctx, &portfoliopb.GetPortfolioSummaryRequest{UserId: userId})
	if err != nil {
		log.Printf("Error checking funds for user %s: %v", userId, err)
		return false
	}

	for _, acc := range resp.Accounts {
		if acc.Type == portfoliopb.AccountType_ACCOUNT_TYPE_INVESTMENT {
			if acc.Balance >= requiredAmount {
				return true
			}
		}
	}
	return false
}

func (e *Engine) checkHoldings(ctx context.Context, userId string, symbol string, requiredQty float64) bool {
	// first get accounts of current user
	resp, err := e.portfolioClient.GetPortfolioSummary(ctx, &portfoliopb.GetPortfolioSummaryRequest{UserId: userId})
	if err != nil {
		log.Printf("Error getting accounts for holding check for user %s: %v", userId, err)
		return false
	}

	// check each account for holdings
	for _, acc := range resp.Accounts {
		if acc.Type == portfoliopb.AccountType_ACCOUNT_TYPE_INVESTMENT {
			hResp, err := e.portfolioClient.GetHolding(ctx, &portfoliopb.GetHoldingRequest{
				AccountId: acc.Id,
				Symbol:    symbol,
			})
			if err == nil && hResp.Holding != nil {
				if hResp.Holding.Quantity >= requiredQty {
					return true
				}
			}
		}
	}
	return false
}

func (e *Engine) publishFilledEvent(order *orderpb.OrderCreatedEvent, fillPrice float64) {
	filledEvent := &orderpb.OrderFilledEvent{
		OrderId:      order.OrderId,
		UserId:       order.UserId,
		Symbol:       order.Symbol,
		Side:         order.Side,
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
