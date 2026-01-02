package api

import (
	"context"
	"encoding/json"
	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/user"
	natsC "fafnir/shared/pkg/nats"
	"fafnir/user-service/internal/db"
	"fafnir/user-service/internal/db/generated"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
)

type UserHandler struct {
	db         *db.Database
	natsClient *natsC.NatsClient
	pb.UnimplementedUserServiceServer
}

func NewUserHandler(database *db.Database, natsClient *natsC.NatsClient) *UserHandler {
	return &UserHandler{
		db:         database,
		natsClient: natsClient,
	}
}

func (h *UserHandler) GetProfileData(ctx context.Context, req *pb.ProfileDataRequest) (*pb.ProfileDataResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return &pb.ProfileDataResponse{
			Data: nil,
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, nil // return nil since this is not a server error (invalid input)
	}

	profileData, err := h.db.GetQueries().GetUserProfileById(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &pb.ProfileDataResponse{
				Data: nil,
				Code: basepb.ErrorCode_NOT_FOUND,
			}, nil
		}

		return nil, err
	}

	return &pb.ProfileDataResponse{
		Data: &pb.ProfileData{
			UserId:    profileData.ID.String(),
			FirstName: profileData.FirstName,
			LastName:  profileData.LastName,
		},
		Code: basepb.ErrorCode_OK,
	}, nil
}

func (h *UserHandler) RegisterSubscribeHandlers() {
	_, err := h.natsClient.QueueSubscribe("users.>", "users-service-main", "users-consumer", h.handleUserEvents)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) handleUserEvents(msg *nats.Msg) {
	var err error

	switch msg.Subject {
	case "users.registered":
		err = h.registerUser(msg)
	case "users.deleted":
		err = h.deleteUser(msg)
	default:
		// ignore events we don't care about
		// we must ack them, otherwise they come back forever
		msg.Ack()
		return
	}

	if err != nil {
		log.Printf("Failed to process %s: %v", msg.Subject, err)
		msg.Nak() // retry later (negative ack)
	} else {
		msg.Ack() // success (acknowledge message)
	}
}

func (h *UserHandler) registerUser(msg *nats.Msg) error {
	var userData struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user registered event: %v", err)
		return nil // don't want to retry unmarshaling errors
	}

	params := generated.InsertUserProfileByIdParams{
		ID:        uuid.MustParse(userData.UserID),
		Email:     userData.Email,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
	}

	_, err := h.db.GetQueries().InsertUserProfileById(context.Background(), params)
	if err != nil {
		log.Printf("Error creating user profile: %v", err)
		return err // want to retry on DB errors
	}

	log.Printf("User profile created for user ID: %s", userData.UserID)
	return nil
}

func (h *UserHandler) deleteUser(msg *nats.Msg) error {
	var userData struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user deleted event: %v", err)
		return nil // don't want to retry unmarshaling errors
	}

	uid := userData.UserID

	if err := h.db.GetQueries().DeleteUserProfileById(context.Background(), uuid.MustParse(userData.UserID)); err != nil {
		log.Printf("Error deleting user profile: %v", err)
		return err // want to retry on DB errors
	}

	log.Printf("User profile deleted for user ID: %s", uid)
	return nil
}
