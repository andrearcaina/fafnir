package api

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GetEmail implements the AuthRequest interface
func (r LoginRequest) GetEmail() string {
	return r.Email
}

// GetPassword implements the AuthRequest interface
func (r LoginRequest) GetPassword() string {
	return r.Password
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GetEmail implements the AuthRequest interface
func (r RegisterRequest) GetEmail() string {
	return r.Email
}

// GetPassword implements the AuthRequest interface
func (r RegisterRequest) GetPassword() string {
	return r.Password
}

type UserInfoRequest struct {
	UserId uuid.UUID `json:"userId"`
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
