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

type UserInfoRequest struct {
	UserId uuid.UUID `json:"userId"`
}

type LoginResponse struct {
	Message  string `json:"message"`
	JwtToken string `json:"jwtToken"`
}

type RegisterResponse struct {
	UserId  uuid.UUID `json:"userId"`
	Message string    `json:"message"`
}

type UserInfoResponse struct {
	UserId uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}
