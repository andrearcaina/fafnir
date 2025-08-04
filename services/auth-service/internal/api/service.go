package api

import (
	"context"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type Code int

type Error error

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

func (s *Service) RegisterUser(ctx context.Context, request RegisterRequest) (RegisterResponse, Code, Error) {
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

	return RegisterResponse{
		UserId:  user.ID,
		Message: "User registered successfully",
	}, http.StatusOK, nil
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, Code, Error) {
	user, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)

	log.Printf("Login attempt for email: %s", request.Email)
	log.Printf("User found: %v", user)

	if err != nil {
		return LoginResponse{
			Message: "Invalid email or password",
		}, http.StatusUnauthorized, err
	}

	if !checkPasswordHash(request.Password, user.PasswordHash) {
		return LoginResponse{
			Message: "Invalid email or password",
		}, http.StatusUnauthorized, nil
	}

	jwtToken, err := s.generateJWT(user.ID)
	if err != nil {
		return LoginResponse{
			Message: "Failed to generate JWT token",
		}, http.StatusInternalServerError, err
	}

	return LoginResponse{
		Message:  "Login successful",
		JwtToken: jwtToken,
	}, http.StatusOK, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID uuid.UUID) (UserInfoResponse, Code, Error) {
	user, err := s.db.GetQueries().GetUserById(ctx, userID)
	if err != nil {
		return UserInfoResponse{
			UserId: uuid.Nil,
			Email:  "",
		}, http.StatusNotFound, err
	}

	return UserInfoResponse{
		UserId: user.ID,
		Email:  user.Email,
	}, http.StatusOK, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Service) generateJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour).Unix(), // token valid for 1 hour
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtKey))
}
