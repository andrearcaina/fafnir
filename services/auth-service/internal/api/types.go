package api

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

type RegisterResponse struct {
	UserID  uuid.UUID `json:"user_id"`
	Message string    `json:"message"`
}
