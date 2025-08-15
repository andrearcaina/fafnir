package api

import (
	"context"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	apperrors "fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/utils"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db     *db.Database
	jwtKey string
}

func NewAuthService(database *db.Database, jwtKey string) *Service {
	return &Service{
		db:     database,
		jwtKey: jwtKey,
	}
}

func (s *Service) RegisterUser(ctx context.Context, request RegisterRequest) (*RegisterResponse, error) {
	_, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)
	if err == nil {
		return nil, apperrors.ConflictError("User already exists").
			WithDetails(fmt.Sprintf("A user with email %s already exists", request.Email))
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.InternalError("Failed to hash password").
			WithDetails("Could not hash the provided password")
	}

	params := generated.RegisterUserParams{
		Email:        request.Email,
		PasswordHash: string(passwordHashBytes),
	}

	user, err := s.db.GetQueries().RegisterUser(ctx, params)
	if err != nil {
		return nil, apperrors.DatabaseError(err).
			WithDetails("Could not create user record")
	}

	return &RegisterResponse{
		UserId:  user.ID,
		Message: "User registered successfully",
	}, nil
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (*LoginResponse, error) {
	user, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, apperrors.NotFoundError("User not found").
			WithDetails(fmt.Sprintf("No user found with email %s", request.Email))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return nil, apperrors.UnauthorizedError().
			WithDetails("The provided email or password is incorrect")
	}

	jwtToken, err := utils.GenerateJWTToken(user.ID, s.jwtKey)
	if err != nil {
		return nil, apperrors.InternalError("Failed to generate JWT token").
			WithDetails("Could not create JWT token for the user")
	}

	csrfToken, err := utils.GenerateCSRFToken(32)
	if err != nil {
		return nil, apperrors.InternalError("Failed to generate CSRF token").
			WithDetails("Could not create CSRF token for the user")
	}

	return &LoginResponse{
		Message:   "Login successful",
		JwtToken:  jwtToken,
		CsrfToken: csrfToken,
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfoResponse, error) {
	user, err := s.db.GetQueries().GetUserById(ctx, userID)
	if err != nil {
		return nil, apperrors.NotFoundError("User not found").
			WithDetails(fmt.Sprintf("No user found with ID %s", userID.String()))
	}

	return &UserInfoResponse{
		UserId: user.ID,
		Email:  user.Email,
	}, nil
}
