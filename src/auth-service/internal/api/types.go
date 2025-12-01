package api

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type UserInfoRequest struct {
	UserId uuid.UUID `json:"userId" validate:"required"`
}

type LoginResponse struct {
	Message   string `json:"message"`
	JwtToken  string `json:"jwtToken"`
	CsrfToken string `json:"csrfToken"`
}

type RegisterResponse struct {
	UserId  uuid.UUID `json:"userId"`
	Message string    `json:"message"`
}

type UserInfoResponse struct {
	UserId uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}
