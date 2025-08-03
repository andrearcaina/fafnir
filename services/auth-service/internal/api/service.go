package api

import (
	"context"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Service struct {
	db *db.Database
}

func NewAuthService(database *db.Database) *Service {
	return &Service{
		db: database,
	}
}

func (s *Service) RegisterUser(ctx context.Context, request RegisterRequest) (RegisterResponse, int, error) {
	_, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)
	if err == nil {
		return RegisterResponse{
			Message: "Email already exists",
		}, http.StatusConflict, err
	}

	passwordHash, err := hashPassword(request.Password)
	if err != nil {
		return RegisterResponse{
			Message: "Failed to hash password",
		}, http.StatusInternalServerError, err
	}

	params := generated.RegisterUserParams{
		Email:        request.Email,
		PasswordHash: passwordHash,
	}

	user, err := s.db.GetQueries().RegisterUser(ctx, params)
	if err != nil {
		return RegisterResponse{
			Message: "Failed to register user",
		}, http.StatusInternalServerError, err
	}

	response := RegisterResponse{
		UserID:  user.ID,
		Message: "User registered successfully",
	}

	return response, http.StatusOK, nil
}

func (s *Service) Login(email, password string) bool {
	// placeholder logic for authentication (will implement real logic with JWT and cookies later)
	return email == "email@email.com" && password == "pass"
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
