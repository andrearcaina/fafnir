package fmp

import (
	"fafnir/stock-service/internal/dto"
	"fmt"

	"resty.dev/v3"
)

type Client struct {
	apiKey string
	client *resty.Client
}

func NewFMPClient(apiKey string) (*Client, error) {
	client := resty.New()
	client.SetBaseURL("https://financialmodelingprep.com/stable")

	if client == nil {
		return nil, fmt.Errorf("failed to create resty client")
	}

	return &Client{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (c *Client) GetStockQuote(symbol string) (*dto.StockQuoteResponse, error) {
	var result []dto.StockQuoteResponse

	_, err := c.client.R().
		SetQueryParams(map[string]string{
			"symbol": symbol,
			"apikey": c.apiKey,
		}).
		SetResult(&result).
		Get("/quote")
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no stock quote")
	}

	return &result[0], nil
}

func (c *Client) GetStockMetadata(symbol string) (*dto.StockMetadataResponse, error) {
	var result []dto.StockMetadataResponse

	_, err := c.client.R().
		SetQueryParams(map[string]string{
			"query":  symbol,
			"apikey": c.apiKey,
		}).
		SetResult(&result).
		Get("/search-symbol")
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no stock metadata")
	}

	return &result[0], nil
}
