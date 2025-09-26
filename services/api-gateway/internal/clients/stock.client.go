package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"

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
