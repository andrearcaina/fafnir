package api

import (
	"context"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	apperrors "fafnir/shared/pkg/errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
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

func (s *Service) RegisterUser(ctx context.Context, request RegisterRequest) (RegisterResponse, error) {
	_, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)
	if err == nil {
		return RegisterResponse{}, apperrors.ConflictError("User already exists").
			WithDetails(fmt.Sprintf("A user with email %s already exists", request.Email))
	}

	passwordHash, err := s.hashPassword(request.Password)
	if err != nil {
		return RegisterResponse{}, apperrors.InternalError("Failed to hash password").
			WithDetails("Could not hash the provided password")
	}

	params := generated.RegisterUserParams{
		Email:        request.Email,
		PasswordHash: passwordHash,
	}

	user, err := s.db.GetQueries().RegisterUser(ctx, params)
	if err != nil {
		return RegisterResponse{}, apperrors.DatabaseError(err).
			WithDetails("Could not create user record")
	}

	return RegisterResponse{
		UserId:  user.ID,
		Message: "User registered successfully",
	}, nil
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {
	user, err := s.db.GetQueries().GetUserByEmail(ctx, request.Email)
	if err != nil {
		return LoginResponse{}, apperrors.NotFoundError("User not found").
			WithDetails(fmt.Sprintf("No user found with email %s", request.Email))
	}

	if !s.checkPasswordHash(request.Password, user.PasswordHash) {
		return LoginResponse{}, apperrors.UnauthorizedError().
			WithDetails("The provided email or password is incorrect")
	}

	jwtToken, err := s.generateJWT(user.ID)
	if err != nil {
		return LoginResponse{}, apperrors.InternalError("Failed to generate JWT token").
			WithDetails("Could not create JWT token for the user")
	}

	return LoginResponse{
		Message:  "Login successful",
		JwtToken: jwtToken,
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID uuid.UUID) (UserInfoResponse, error) {
	user, err := s.db.GetQueries().GetUserById(ctx, userID)
	if err != nil {
		return UserInfoResponse{}, apperrors.NotFoundError("User not found").
			WithDetails(fmt.Sprintf("No user found with ID %s", userID.String()))
	}

	return UserInfoResponse{
		UserId: user.ID,
		Email:  user.Email,
	}, nil
}

func (s *Service) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *Service) checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Service) parseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.TokenError("Unexpected signing method used").
				WithDetails(fmt.Sprintf("Expected signing method HS256, got %s", token.Header["alg"]))
		}
		return []byte(s.jwtKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, apperrors.TokenError("Invalid token").
			WithDetails("The provided JWT token is invalid or expired")
	}

	return token, nil
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
