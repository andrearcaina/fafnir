package api

import (
	"context"
	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/logger"
)

type StockHandler struct {
	stockService *Service
	logger       *logger.Logger
	pb.UnimplementedStockServiceServer
}

func NewStockHandler(stockService *Service, logger *logger.Logger) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		logger:       logger,
	}
}

// GetStockMetadata implements the gRPC GetStockMetadata method
func (h *StockHandler) GetStockMetadata(ctx context.Context, req *pb.GetStockMetadataRequest) (*pb.GetStockMetadataResponse, error) {
	metadata, err := h.stockService.GetStockMetadata(ctx, req.Symbol)
	if err != nil {
		// if err is an app error with code bad request, return INVALID_ARGUMENT without error
		if errors.Is(err, errors.BadRequestError("")) {
			return &pb.GetStockMetadataResponse{
				Data: nil,
				Code: basepb.ErrorCode_INVALID_ARGUMENT,
			}, nil
		} else if errors.Is(err, errors.InternalError("")) {
			return &pb.GetStockMetadataResponse{
				Data: nil,
				Code: basepb.ErrorCode_INTERNAL,
			}, nil
		}

		// for other errors, just return the error
		return nil, err
	}

	return &pb.GetStockMetadataResponse{
		Data: &pb.StockMetadata{
			Symbol:           metadata.Symbol,
			Name:             metadata.Name,
			Currency:         metadata.Currency,
			Exchange:         metadata.Exchange,
			ExchangeFullName: metadata.ExchangeFullName,
		},
		Code: basepb.ErrorCode_OK,
	}, nil
}

// GetStockQuote implements the gRPC GetStockQuote method
func (h *StockHandler) GetStockQuote(ctx context.Context, req *pb.GetStockQuoteRequest) (*pb.GetStockQuoteResponse, error) {
	quote, err := h.stockService.GetStockQuote(ctx, req.Symbol)
	if err != nil {
		if errors.Is(err, errors.BadRequestError("")) {
			return &pb.GetStockQuoteResponse{
				Data: nil,
				Code: basepb.ErrorCode_INVALID_ARGUMENT,
			}, nil
		} else if errors.Is(err, errors.InternalError("")) {
			return &pb.GetStockQuoteResponse{
				Data: nil,
				Code: basepb.ErrorCode_INTERNAL,
			}, nil
		}

		return nil, err
	}

	return &pb.GetStockQuoteResponse{
		Data: &pb.StockQuote{
			Symbol:        quote.Symbol,
			LastPrice:     quote.LastPrice,
			OpenPrice:     quote.OpenPrice,
			PreviousClose: quote.PreviousClose,
			DayLow:        quote.DayLow,
			DayHigh:       quote.DayHigh,
			YearLow:       quote.YearLow,
			YearHigh:      quote.YearHigh,
			Volume:        quote.Volume,
			MarketCap:     quote.MarketCap,
			Change:        quote.Change,
			ChangePct:     quote.ChangePct,
		},
		Code: basepb.ErrorCode_OK,
	}, nil
}

// GetStockQuoteBatch implements the gRPC GetStockQuoteBatch method
func (h *StockHandler) GetStockQuoteBatch(ctx context.Context, req *pb.GetStockQuoteBatchRequest) (*pb.GetStockQuoteBatchResponse, error) {
	quotes, err := h.stockService.GetStockQuoteBatch(ctx, req.Symbols)
	if err != nil {
		if errors.Is(err, errors.BadRequestError("")) {
			return &pb.GetStockQuoteBatchResponse{
				Data: nil,
				Code: basepb.ErrorCode_INVALID_ARGUMENT,
			}, nil
		} else if errors.Is(err, errors.InternalError("")) {
			return &pb.GetStockQuoteBatchResponse{
				Data: nil,
				Code: basepb.ErrorCode_INTERNAL,
			}, nil
		}

		return nil, err
	}

	var pbQuotes []*pb.StockQuote
	for _, quote := range quotes {
		pbQuotes = append(pbQuotes, &pb.StockQuote{
			Symbol:        quote.Symbol,
			LastPrice:     quote.LastPrice,
			OpenPrice:     quote.OpenPrice,
			PreviousClose: quote.PreviousClose,
			DayLow:        quote.DayLow,
			DayHigh:       quote.DayHigh,
			YearLow:       quote.YearLow,
			YearHigh:      quote.YearHigh,
			Volume:        quote.Volume,
			MarketCap:     quote.MarketCap,
			Change:        quote.Change,
			ChangePct:     quote.ChangePct,
		})
	}

	return &pb.GetStockQuoteBatchResponse{
		Data: pbQuotes,
		Code: basepb.ErrorCode_OK,
	}, nil
}

// getStockHistoricalData implements the gRPC GetStockHistoricalData method
func (h *StockHandler) GetStockHistoricalData(ctx context.Context, req *pb.GetStockHistoricalDataRequest) (*pb.GetStockHistoricalDataResponse, error) {
	historicalData, err := h.stockService.GetStockHistoricalData(ctx, req.Symbol, req.Period)
	if err != nil {
		if errors.Is(err, errors.BadRequestError("")) {
			return &pb.GetStockHistoricalDataResponse{
				Data: nil,
				Code: basepb.ErrorCode_INVALID_ARGUMENT,
			}, nil
		} else if errors.Is(err, errors.InternalError("")) {
			return &pb.GetStockHistoricalDataResponse{
				Data: nil,
				Code: basepb.ErrorCode_INTERNAL,
			}, nil
		}

		return nil, err
	}

	var pbStockHistoricalData []*pb.StockHistoricalData
	for _, stockData := range historicalData {
		pbStockHistoricalData = append(pbStockHistoricalData, &pb.StockHistoricalData{
			Symbol:     stockData.Symbol,
			Date:       stockData.Date,
			OpenPrice:  stockData.OpenPrice,
			HighPrice:  stockData.HighPrice,
			LowPrice:   stockData.LowPrice,
			ClosePrice: stockData.ClosePrice,
			Volume:     stockData.Volume,
			Change:     stockData.Change,
			ChangePct:  stockData.ChangePct,
		})
	}

	return &pb.GetStockHistoricalDataResponse{
		Data: pbStockHistoricalData,
		Code: basepb.ErrorCode_OK,
	}, nil
}
