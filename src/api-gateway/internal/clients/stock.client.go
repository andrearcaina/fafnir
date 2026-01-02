package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	"strings"

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

func (c *StockClient) GetStockMetadata(ctx context.Context, symbol string) (model.StockMetadataResponse, error) {
	req := &pb.GetStockMetadataRequest{
		Symbol: symbol,
	}

	resp, err := c.client.GetStockMetadata(ctx, req)
	if err != nil {
		return model.StockMetadataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, err
	}

	if resp.GetCode() == basepb.ErrorCode_INVALID_ARGUMENT {
		return model.StockMetadataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	} else if resp.GetCode() == basepb.ErrorCode_INTERNAL {
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
			Code: resp.GetCode().String(),
		}, err
	}

	if resp.GetCode() == basepb.ErrorCode_INVALID_ARGUMENT {
		return model.StockQuoteResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	} else if resp.GetCode() == basepb.ErrorCode_INTERNAL {
		return model.StockQuoteResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	return model.StockQuoteResponse{
		Data: &model.StockPriceData{
			Symbol:             resp.GetData().GetSymbol(),
			Price:              resp.GetData().GetLastPrice(),
			Open:               resp.GetData().GetOpenPrice(),
			PreviousClose:      resp.GetData().GetPreviousClose(),
			DayLow:             resp.GetData().GetDayLow(),
			DayHigh:            resp.GetData().GetDayHigh(),
			YearLow:            resp.GetData().GetYearLow(),
			YearHigh:           resp.GetData().GetYearHigh(),
			Volume:             resp.GetData().GetVolume(),
			MarketCap:          resp.GetData().GetMarketCap(),
			PriceChange:        resp.GetData().GetChange(),
			PriceChangePercent: resp.GetData().GetChangePct(),
		},
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
			Code: resp.GetCode().String(),
		}, err
	}

	if resp.GetCode() == basepb.ErrorCode_INVALID_ARGUMENT {
		return model.StockQuoteBatchResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	} else if resp.GetCode() == basepb.ErrorCode_INTERNAL {
		return model.StockQuoteBatchResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	var stockPrices []*model.StockPriceData
	for _, stock := range resp.GetData() {
		stockPrices = append(stockPrices, &model.StockPriceData{
			Symbol:             stock.GetSymbol(),
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
		})
	}

	return model.StockQuoteBatchResponse{
		Data: stockPrices,
		Code: resp.GetCode().String(),
	}, nil
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
			Code: resp.GetCode().String(),
		}, err
	}

	if resp.GetCode() == basepb.ErrorCode_INVALID_ARGUMENT {
		return model.StockHistoricalDataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	} else if resp.GetCode() == basepb.ErrorCode_INTERNAL {
		return model.StockHistoricalDataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	var historicalData []*model.StockHistoricalData
	for _, stockData := range resp.GetData() {
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
