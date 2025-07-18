package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LoginRequest struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type AuthClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *AuthClient) Login(ctx context.Context, user, pass string) (*LoginResponse, error) {
	reqBody := LoginRequest{
		User: user,
		Pass: pass,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := fmt.Sprintf("%s/auth/login", strings.TrimRight(c.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request to auth service: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth service response: %w", err)
	}

	loginResp.Code = resp.StatusCode

	return &loginResp, nil
}
