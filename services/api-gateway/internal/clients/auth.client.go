package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	"fmt"
	"log"
	"resty.dev/v3"
)

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

func (c *AuthClient) Register(ctx context.Context, request model.RegisterRequest) (*model.RegisterResponse, error) {
	log.Printf("Registering user with email %v\n", request)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&model.RegisterResponse{}).
		SetError(&model.RegisterResponse{}).
		Post("/auth/register")

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	var registerResponse model.RegisterResponse
	statusCode := int32(resp.StatusCode())
	registerResponse.Code = &statusCode

	if resp.IsError() {
		registerResponse.Message = resp.Error().(*model.RegisterResponse).Message
	} else if resp.IsSuccess() {
		registerResponse.Message = resp.Result().(*model.RegisterResponse).Message
	}

	return &registerResponse, nil
}

func (c *AuthClient) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, error) {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&model.LoginResponse{}).
		SetError(&model.LoginResponse{}).
		Post("/auth/login")

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	var loginResponse model.LoginResponse
	statusCode := int32(resp.StatusCode())
	loginResponse.Code = &statusCode

	if resp.IsError() {
		loginResponse.Message = resp.Error().(*model.LoginResponse).Message
	} else if resp.IsSuccess() {
		loginResponse.Message = resp.Result().(*model.LoginResponse).Message
	}

	return &loginResponse, nil
}
