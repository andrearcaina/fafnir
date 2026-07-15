package clients

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fafnir/api-gateway/graph/model"

	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/stock"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StockClient struct {
	conn   *grpc.ClientConn
	client pb.StockServiceClient
}

func NewStockClient(address string) *StockClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil
	}

	client := pb.NewStockServiceClient(conn)

	return &StockClient{
		conn:   conn,
		client: client,
	}
}

func (c *StockClient) SearchStocks(ctx context.Context, query string, limit int) ([]*model.StockSearchResult, error) {
	resp, err := c.client.SearchStocks(ctx, &pb.SearchStocksRequest{
		Query: query,
		Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	if resp.GetCode() != basepb.ErrorCode_OK {
		return nil, fmt.Errorf("stock search returned %s", resp.GetCode().String())
	}

	results := make([]*model.StockSearchResult, 0, len(resp.Data))
	for _, result := range resp.Data {
		if result == nil {
			continue
		}

		results = append(results, &model.StockSearchResult{
			Symbol:           result.Symbol,
			Name:             result.Name,
			Exchange:         result.Exchange,
			ExchangeFullName: result.ExchangeFullName,
			InstrumentType:   result.InstrumentType,
		})
	}

	return results, nil
}

func (c *StockClient) GetStockMetadata(ctx context.Context, symbol string) (model.StockMetadataResponse, error) {
	req := &pb.GetStockMetadataRequest{
		Symbol: symbol,
	}

	resp, err := c.client.GetStockMetadata(ctx, req)
	if err != nil {
		return model.StockMetadataResponse{
			Data: nil,
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	if resp.GetCode() != basepb.ErrorCode_OK || resp.GetData() == nil {
		return model.StockMetadataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	return model.StockMetadataResponse{
		Data: &model.StockData{
			Symbol:           resp.GetData().GetSymbol(),
			Name:             resp.GetData().GetName(),
			Currency:         resp.GetData().GetCurrency(),
			Exchange:         resp.GetData().GetExchange(),
			ExchangeFullName: resp.GetData().GetExchangeFullName(),
			InstrumentType:   resp.GetData().GetInstrumentType(),
		},
		Code: resp.GetCode().String(),
	}, nil
}

func (c *StockClient) GetStockQuote(ctx context.Context, symbol string) (model.StockQuoteResponse, error) {
	req := &pb.GetStockQuoteRequest{
		Symbol: symbol,
	}

	resp, err := c.client.GetStockQuote(ctx, req)
	if err != nil {
		return model.StockQuoteResponse{
			Data: nil,
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	if resp.GetCode() != basepb.ErrorCode_OK || resp.GetData() == nil {
		return model.StockQuoteResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	return model.StockQuoteResponse{
		Data: quoteToModel(resp.GetData()),
		Code: resp.GetCode().String(),
	}, nil
}

func (c *StockClient) GetStockQuoteBatch(ctx context.Context, symbols []string) (model.StockQuoteBatchResponse, error) {
	req := &pb.GetStockQuoteBatchRequest{
		Symbols: symbols,
	}

	resp, err := c.client.GetStockQuoteBatch(ctx, req)
	if err != nil {
		return model.StockQuoteBatchResponse{
			Data: nil,
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	if resp.GetCode() != basepb.ErrorCode_OK {
		return model.StockQuoteBatchResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	var stockPrices []*model.StockPriceData
	for _, stock := range resp.GetData() {
		if mapped := quoteToModel(stock); mapped != nil {
			stockPrices = append(stockPrices, mapped)
		}
	}

	return model.StockQuoteBatchResponse{
		Data: stockPrices,
		Code: resp.GetCode().String(),
	}, nil
}

func quoteToModel(stock *pb.StockQuote) *model.StockPriceData {
	if stock == nil {
		return nil
	}

	asOf := ""
	if stock.AsOf != nil {
		asOf = stock.AsOf.AsTime().Format(time.RFC3339)
	}

	return &model.StockPriceData{
		Symbol:             stock.GetSymbol(),
		Currency:           stock.GetCurrency(),
		Price:              stock.GetLastPrice(),
		Open:               stock.GetOpenPrice(),
		PreviousClose:      stock.GetPreviousClose(),
		DayLow:             stock.GetDayLow(),
		DayHigh:            stock.GetDayHigh(),
		YearLow:            stock.GetYearLow(),
		YearHigh:           stock.GetYearHigh(),
		Volume:             stock.GetVolume(),
		MarketCap:          stock.GetMarketCap(),
		PriceChange:        stock.GetChange(),
		PriceChangePercent: stock.GetChangePct(),
		Source:             stock.GetSource(),
		AsOf:               asOf,
		MarketState:        stock.GetMarketState(),
	}
}

func (c *StockClient) GetStockHistoricalData(ctx context.Context, symbol string, period string) (model.StockHistoricalDataResponse, error) {
	req := &pb.GetStockHistoricalDataRequest{
		Symbol: symbol,
		Period: strings.ToUpper(period),
	}

	resp, err := c.client.GetStockHistoricalData(ctx, req)
	if err != nil {
		return model.StockHistoricalDataResponse{
			Data: nil,
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	if resp.GetCode() != basepb.ErrorCode_OK {
		return model.StockHistoricalDataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	var historicalData []*model.StockHistoricalData
	for _, stockData := range resp.GetData() {
		if stockData == nil {
			continue
		}
		historicalData = append(historicalData, &model.StockHistoricalData{
			Symbol:             stockData.GetSymbol(),
			Date:               stockData.GetDate(),
			Open:               stockData.GetOpenPrice(),
			Close:              stockData.GetClosePrice(),
			High:               stockData.GetHighPrice(),
			Low:                stockData.GetLowPrice(),
			Volume:             stockData.GetVolume(),
			PriceChange:        stockData.GetChange(),
			PriceChangePercent: stockData.GetChangePct(),
		})
	}

	return model.StockHistoricalDataResponse{
		Data: historicalData,
		Code: resp.GetCode().String(),
	}, nil
}
