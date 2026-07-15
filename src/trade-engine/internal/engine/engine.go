package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	basepb "fafnir/shared/pb/base"
	orderpb "fafnir/shared/pb/order"
	portfoliopb "fafnir/shared/pb/portfolio"
	stockpb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/logger"
	natsclient "fafnir/shared/pkg/nats"
	"fafnir/shared/pkg/redis"
	"fafnir/trade-engine/internal/cache"
	"fafnir/trade-engine/internal/config"
	"fafnir/trade-engine/internal/fx"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	orderPollInterval = 5 * time.Second
	requestTimeout    = 10 * time.Second
	retryDelay        = 2 * time.Second
)

type Engine struct {
	natsClient      *natsclient.NatsClient
	stockClient     stockpb.StockServiceClient
	portfolioClient portfoliopb.PortfolioServiceClient
	fxProvider      fx.Provider
	orderBook       *cache.OrderBook
	stockConn       *grpc.ClientConn
	portfolioConn   *grpc.ClientConn
	redisClient     *redis.Cache
	stopCh          chan struct{}
	stopOnce        sync.Once
	logger          *logger.Logger
}

func NewEngine(cfg *config.Config, log *logger.Logger) (*Engine, error) {
	log.Info(context.Background(), "Engine connecting to NATS", "url", cfg.NATS.URL)
	nc, err := natsclient.New(cfg.NATS.URL, log)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	stockConn, err := grpc.NewClient(cfg.StockService.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("connect to stock service: %w", err)
	}

	portfolioConn, err := grpc.NewClient(cfg.Portfolio.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		stockConn.Close()
		nc.Close()
		return nil, fmt.Errorf("connect to portfolio service: %w", err)
	}

	redisClient, err := redis.New(cfg.Cache, log)
	if err != nil {
		portfolioConn.Close()
		stockConn.Close()
		nc.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}

	return &Engine{
		natsClient:      nc,
		stockClient:     stockpb.NewStockServiceClient(stockConn),
		portfolioClient: portfoliopb.NewPortfolioServiceClient(portfolioConn),
		fxProvider:      fx.NewFrankfurter(cfg.FX.BaseURL, cfg.FX.Timeout, cfg.FX.TTL),
		orderBook:       cache.NewOrderBook(redisClient),
		stockConn:       stockConn,
		portfolioConn:   portfolioConn,
		redisClient:     redisClient,
		stopCh:          make(chan struct{}),
		logger:          log,
	}, nil
}

func (e *Engine) Start() error {
	if err := e.subscribeToCreatedOrders(); err != nil {
		return fmt.Errorf("subscribe to created orders: %w", err)
	}
	if err := e.subscribeToCancelledOrders(); err != nil {
		return fmt.Errorf("subscribe to cancelled orders: %w", err)
	}

	go e.pollOrders()
	<-e.stopCh

	return nil
}

func (e *Engine) Stop() error {
	var closeErr error
	e.stopOnce.Do(func() {
		close(e.stopCh)
		e.natsClient.Close()

		closeErr = errors.Join(
			e.fxProvider.Close(),
			e.stockConn.Close(),
			e.portfolioConn.Close(),
			e.redisClient.Close(),
		)
	})

	return closeErr
}

func (e *Engine) subscribeToCreatedOrders() error {
	_, err := e.natsClient.QueueSubscribe("orders.created", "trade-engine", "trade-engine-durable", func(msg *nats.Msg) {
		var event orderpb.OrderCreatedEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			e.logger.Error(context.Background(), "Discarding malformed order event", "error", err)
			_ = msg.Term()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		if err := e.processOrder(ctx, &event); err != nil {
			e.logger.Error(ctx, "Order processing failed; scheduling retry", "order_id", event.OrderId, "error", err)
			_ = msg.NakWithDelay(retryDelay)
			return
		}

		_ = msg.Ack()
	})

	return err
}

func (e *Engine) subscribeToCancelledOrders() error {
	_, err := e.natsClient.QueueSubscribe("orders.cancelled", "trade-engine", "trade-engine-cancelled", func(msg *nats.Msg) {
		var event orderpb.OrderCancelledEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			e.logger.Error(context.Background(), "Discarding malformed cancellation event", "error", err)
			_ = msg.Term()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		if err := e.orderBook.Remove(ctx, event.Symbol, event.OrderId); err != nil {
			e.logger.Error(ctx, "Failed to remove cancelled order", "order_id", event.OrderId, "error", err)
			_ = msg.NakWithDelay(retryDelay)
			return
		}

		_ = msg.Ack()
	})

	return err
}

func (e *Engine) processOrder(ctx context.Context, order *orderpb.OrderCreatedEvent) error {
	if err := validateOrder(order); err != nil {
		return e.publishRejectedEvent(ctx, order, err.Error())
	}

	quote, err := e.getQuote(ctx, order.Symbol)
	if err != nil {
		return err
	}

	if evaluateOrder(order, quote.LastPrice) == decisionWait {
		if err := e.orderBook.Add(ctx, order); err != nil {
			return fmt.Errorf("queue limit order: %w", err)
		}

		e.logger.Info(ctx, "Limit order queued", "order_id", order.OrderId, "symbol", order.Symbol, "limit_price", order.Price)
		return nil
	}

	return e.executeAtPrice(ctx, order, quote.LastPrice)
}

func (e *Engine) pollOrders() {
	ticker := time.NewTicker(orderPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.pollOnce()
		}
	}
}

func (e *Engine) pollOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	symbols, err := e.orderBook.Symbols(ctx)
	if err != nil {
		e.logger.Error(ctx, "Failed to list queued order symbols", "error", err)
		return
	}
	if len(symbols) == 0 {
		return
	}

	resp, err := e.stockClient.GetStockQuoteBatch(ctx, &stockpb.GetStockQuoteBatchRequest{Symbols: symbols})
	if err != nil {
		e.logger.Error(ctx, "Failed to fetch quotes for queued orders", "error", err)
		return
	}
	if resp.Code != basepb.ErrorCode_OK || len(resp.Data) == 0 {
		e.logger.Error(ctx, "Stock service returned no usable batch quotes", "code", resp.Code.String())
		return
	}

	for _, quote := range resp.Data {
		orders, err := e.orderBook.ClaimMatched(ctx, quote.Symbol, quote.LastPrice)
		if err != nil {
			e.logger.Error(ctx, "Failed to claim matching limit orders", "symbol", quote.Symbol, "error", err)
			continue
		}

		for _, order := range orders {
			if err := e.executeAtPrice(ctx, order, quote.LastPrice); err != nil {
				e.logger.Error(ctx, "Matched limit order failed; returning it to the queue", "order_id", order.OrderId, "error", err)
				if addErr := e.orderBook.Add(ctx, order); addErr != nil {
					e.logger.Error(ctx, "Failed to requeue limit order", "order_id", order.OrderId, "error", addErr)
				}
			}
		}
	}
}

func (e *Engine) executeAtPrice(ctx context.Context, order *orderpb.OrderCreatedEvent, fillPrice float64) error {
	if !positiveFinite(fillPrice) {
		return fmt.Errorf("fill price must be greater than zero")
	}

	metadata, err := e.getMetadata(ctx, order.Symbol)
	if err != nil {
		return err
	}
	if !isTradableInstrument(metadata.InstrumentType) {
		return e.publishRejectedEvent(ctx, order, fmt.Sprintf("%s instruments are not supported for trading", metadata.InstrumentType))
	}

	account, err := e.getInvestmentAccount(ctx, order.UserId)
	if err != nil {
		return e.publishRejectedEvent(ctx, order, "No investment account found")
	}

	accountCurrency, err := currencyCode(account.Currency)
	if err != nil {
		return e.publishRejectedEvent(ctx, order, err.Error())
	}

	exchangeRate, err := e.fxProvider.Rate(ctx, metadata.Currency, accountCurrency)
	if err != nil {
		return fmt.Errorf("get %s/%s exchange rate: %w", metadata.Currency, accountCurrency, err)
	}
	if !positiveFinite(exchangeRate) {
		return fmt.Errorf("get %s/%s exchange rate: provider returned an invalid rate", metadata.Currency, accountCurrency)
	}

	settlementAmount := fillPrice * order.Quantity * exchangeRate
	if !positiveFinite(settlementAmount) {
		return fmt.Errorf("calculate settlement amount: result is invalid")
	}
	if order.Side == orderpb.OrderSide_ORDER_SIDE_BUY && account.Balance < settlementAmount {
		return e.publishRejectedEvent(ctx, order, fmt.Sprintf("Insufficient funds: need %.2f %s", settlementAmount, accountCurrency))
	}
	if order.Side == orderpb.OrderSide_ORDER_SIDE_SELL {
		sufficient, err := e.hasSufficientHoldings(ctx, account.Id, order.Symbol, order.Quantity)
		if err != nil {
			return err
		}
		if !sufficient {
			return e.publishRejectedEvent(ctx, order, "Insufficient holdings")
		}
	}

	return e.publishFilledEvent(ctx, order, fillPrice, exchangeRate, settlementAmount, accountCurrency)
}

func (e *Engine) getQuote(ctx context.Context, symbol string) (*stockpb.StockQuote, error) {
	resp, err := e.stockClient.GetStockQuote(ctx, &stockpb.GetStockQuoteRequest{Symbol: symbol})
	if err != nil {
		return nil, fmt.Errorf("get quote for %s: %w", symbol, err)
	}
	if resp.Code != basepb.ErrorCode_OK || resp.Data == nil || resp.Data.LastPrice <= 0 {
		return nil, fmt.Errorf("get quote for %s: stock service returned %s", symbol, resp.Code.String())
	}

	return resp.Data, nil
}

func (e *Engine) getMetadata(ctx context.Context, symbol string) (*stockpb.StockMetadata, error) {
	resp, err := e.stockClient.GetStockMetadata(ctx, &stockpb.GetStockMetadataRequest{Symbol: symbol})
	if err != nil {
		return nil, fmt.Errorf("get metadata for %s: %w", symbol, err)
	}
	if resp.Code != basepb.ErrorCode_OK || resp.Data == nil || resp.Data.Currency == "" || resp.Data.InstrumentType == "" {
		return nil, fmt.Errorf("get metadata for %s: stock service returned %s", symbol, resp.Code.String())
	}

	return resp.Data, nil
}

func (e *Engine) getInvestmentAccount(ctx context.Context, userID string) (*portfoliopb.Account, error) {
	resp, err := e.portfolioClient.GetPortfolioSummary(ctx, &portfoliopb.GetPortfolioSummaryRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("get portfolio summary: %w", err)
	}
	if resp.GetCode() != basepb.ErrorCode_OK {
		return nil, fmt.Errorf("get portfolio summary: portfolio service returned %s", resp.GetCode().String())
	}

	for _, account := range resp.Accounts {
		if account.Type == portfoliopb.AccountType_ACCOUNT_TYPE_INVESTMENT {
			return account, nil
		}
	}

	return nil, fmt.Errorf("no investment account found for user %s", userID)
}

func (e *Engine) hasSufficientHoldings(ctx context.Context, accountID string, symbol string, quantity float64) (bool, error) {
	resp, err := e.portfolioClient.GetHolding(ctx, &portfoliopb.GetHoldingRequest{
		AccountId: accountID,
		Symbol:    symbol,
	})
	if err != nil {
		return false, fmt.Errorf("get %s holding: %w", symbol, err)
	}
	if resp.GetCode() == basepb.ErrorCode_NOT_FOUND {
		return false, nil
	}
	if resp.GetCode() != basepb.ErrorCode_OK {
		return false, fmt.Errorf("get %s holding: portfolio service returned %s", symbol, resp.GetCode().String())
	}
	if resp.Holding == nil {
		return false, fmt.Errorf("get %s holding: portfolio service returned no holding", symbol)
	}

	return resp.Holding.Quantity >= quantity, nil
}

func (e *Engine) publishRejectedEvent(ctx context.Context, order *orderpb.OrderCreatedEvent, reason string) error {
	event := &orderpb.OrderRejectedEvent{
		OrderId:    order.OrderId,
		UserId:     order.UserId,
		Symbol:     order.Symbol,
		Reason:     reason,
		RejectedAt: timestamppb.Now(),
	}

	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal rejected event: %w", err)
	}
	if _, err := e.natsClient.PublishWithID("orders.rejected", order.OrderId+":rejected", data); err != nil {
		return fmt.Errorf("publish rejected event: %w", err)
	}

	e.logger.Info(ctx, "Order rejected", "order_id", order.OrderId, "reason", reason)
	return nil
}

func (e *Engine) publishFilledEvent(ctx context.Context, order *orderpb.OrderCreatedEvent, fillPrice float64, exchangeRate float64, settlementAmount float64, settlementCurrency string) error {
	event := &orderpb.OrderFilledEvent{
		OrderId:            order.OrderId,
		UserId:             order.UserId,
		Symbol:             order.Symbol,
		Side:               order.Side,
		FillQuantity:       order.Quantity,
		FillPrice:          fillPrice,
		ExchangeRate:       exchangeRate,
		SettlementAmount:   settlementAmount,
		SettlementCurrency: settlementCurrency,
		FilledAt:           timestamppb.Now(),
	}

	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal filled event: %w", err)
	}
	if _, err := e.natsClient.PublishWithID("orders.filled", order.OrderId+":filled", data); err != nil {
		return fmt.Errorf("publish filled event: %w", err)
	}

	e.logger.Info(ctx, "Order filled", "order_id", order.OrderId, "fill_price", fillPrice, "settlement_amount", settlementAmount, "settlement_currency", settlementCurrency)
	return nil
}
