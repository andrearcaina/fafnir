package api

import (
	"context"
	"encoding/json"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	apperrors "fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/nats"
	"fafnir/shared/pkg/utils"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db     *db.Database
	nats   *nats.NatsClient
	jwtKey string
}

func NewAuthService(database *db.Database, natsClient *nats.NatsClient, jwtKey string) *Service {
	return &Service{
		db:     database,
		nats:   natsClient,
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

	// create user in postgres database
	user, err := s.db.GetQueries().RegisterUser(ctx, params)
	if err != nil {
		return nil, apperrors.DatabaseError(err).
			WithDetails("Could not create user record")
	}

	// then publish "user.registered" event to Nats
	publishPayload, err := json.Marshal(map[string]string{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"first_name": request.FirstName,
		"last_name":  request.LastName,
	})
	if err != nil {
		return nil, apperrors.InternalError("Failed to marshal event payload").
			WithDetails("Could not marshal user registered event payload")
	}

	// publish to NATS server so that other services can consume the event (e.g. user-service)
	if err := s.nats.Publish("user.registered", publishPayload); err != nil {
		return nil, apperrors.InternalError("Failed to publish user registered event").
			WithDetails("Could not publish user registered event to NATS")
	}

	// return successful response
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

func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	user, err := s.db.GetQueries().GetUserById(ctx, userID)
	if err != nil {
		return apperrors.NotFoundError("User not found").
			WithDetails(fmt.Sprintf("No user found with ID %s", userID.String()))
	}

	// delete user from postgres database
	if err := s.db.GetQueries().DeleteUserById(ctx, user.ID); err != nil {
		return apperrors.DatabaseError(err).
			WithDetails("Could not delete user record")
	}

	// then publish "user.deleted" event to Nats
	publishPayload, err := json.Marshal(map[string]string{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
	if err != nil {
		return apperrors.InternalError("Failed to marshal event payload").
			WithDetails("Could not marshal user deleted event payload")
	}

	// publish to NATS server so that other services can consume the event (e.g. user-service)
	if err := s.nats.Publish("user.deleted", publishPayload); err != nil {
		return apperrors.InternalError("Failed to publish user deleted event").
			WithDetails("Could not publish user deleted event to NATS")
	}

	return nil
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
