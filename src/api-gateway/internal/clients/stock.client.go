package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	"strings"

	"resty.dev/v3"
)

type StockClient struct {
	BaseURL string
	Client  *resty.Client
}

func NewStockClient(baseURL string) *StockClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	return &StockClient{
		BaseURL: baseURL,
		Client:  client,
	}
}

func (c *StockClient) GetStockMetadata(ctx context.Context, symbol string) (model.StockMetadataResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetPathParam("symbol", symbol).
		SetResult(&model.StockData{}).
		SetError(&model.StockData{}).
		Get("/metadata/{symbol}")

	if err != nil {
		return model.StockMetadataResponse{}, err
	}

	var response model.StockMetadataResponse

	response.Code = int32(resp.StatusCode())

	if resp.IsError() {
		response.Data = nil
	}
	if resp.IsSuccess() {
		response.Data = resp.Result().(*model.StockData)
	}

	return response, nil
}

func (c *StockClient) GetStockQuote(ctx context.Context, symbol string) (model.StockQuoteResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetPathParam("symbol", symbol).
		SetResult(&model.StockPriceData{}).
		SetError(&model.StockPriceData{}).
		Get("/quote/{symbol}")

	if err != nil {
		return model.StockQuoteResponse{}, err
	}

	var response model.StockQuoteResponse

	response.Code = int32(resp.StatusCode())

	if resp.IsError() {
		response.Data = nil
	}
	if resp.IsSuccess() {
		response.Data = resp.Result().(*model.StockPriceData)
	}

	return response, nil
}

func (c *StockClient) GetStockQuoteBatch(ctx context.Context, symbols []string) (model.StockQuoteBatchResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetQueryParam("symbols", strings.Join(symbols, ",")).
		SetResult(&[]*model.StockPriceData{}).
		SetError(&[]*model.StockPriceData{}).
		Get("/quote/batch")

	if err != nil {
		return model.StockQuoteBatchResponse{}, err
	}

	var response model.StockQuoteBatchResponse

	response.Code = int32(resp.StatusCode())

	if resp.IsError() {
		response.Data = nil
	}
	if resp.IsSuccess() {
		response.Data = *resp.Result().(*[]*model.StockPriceData)
	}

	return response, nil
}

func (c *StockClient) GetStockHistoricalData(ctx context.Context, symbol string, period string) (model.StockHistoricalDataResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetPathParam("symbol", symbol).
		SetPathParam("period", period).
		SetResult(&[]*model.StockHistoricalData{}).
		SetError(&[]*model.StockHistoricalData{}).
		Get("/historical/{symbol}/{period}")

	if err != nil {
		return model.StockHistoricalDataResponse{}, err
	}

	var response model.StockHistoricalDataResponse

	response.Code = int32(resp.StatusCode())

	if resp.IsError() {
		response.Data = nil
	}
	if resp.IsSuccess() {
		response.Data = *resp.Result().(*[]*model.StockHistoricalData)
	}

	return response, nil
}
