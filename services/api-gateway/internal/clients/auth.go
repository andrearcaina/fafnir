package clients

import (
	"context"
	"fmt"
	"resty.dev/v3"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type AuthClient struct {
	BaseURL string
	Client  *resty.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	return &AuthClient{
		BaseURL: baseURL,
		Client:  client,
	}
}

func (c *AuthClient) Login(ctx context.Context, user, pass string) (*LoginResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(LoginRequest{
			Username: user,
			Password: pass,
		}).
		SetResult(&LoginResponse{}).
		SetError(&LoginResponse{}).
		Post("/auth/login")

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	var loginResponse LoginResponse
	loginResponse.Code = resp.StatusCode()

	if resp.IsError() {
		loginResponse.Message = resp.Error().(*LoginResponse).Message
	} else if resp.IsSuccess() {
		loginResponse.Message = resp.Result().(*LoginResponse).Message
	}

	return &loginResponse, nil
}
