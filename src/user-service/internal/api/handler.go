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
			UserId:    "",
			FirstName: "",
			LastName:  "",
			Code:      basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	profileData, err := h.db.GetQueries().GetUserProfileById(ctx, userID)
	if err != nil {
		return &pb.ProfileDataResponse{
			UserId:    userID.String(),
			FirstName: "",
			LastName:  "",
			Code:      basepb.ErrorCode_NOT_FOUND,
		}, err
	}

	return &pb.ProfileDataResponse{
		UserId:    userID.String(),
		FirstName: profileData.FirstName,
		LastName:  profileData.LastName,
		Code:      basepb.ErrorCode_OK,
	}, nil
}

func (h *UserHandler) RegisterSubscribeHandlers() {
	_, err := h.natsClient.Subscribe("user.registered", h.registerUser)
	if err != nil {
		log.Fatal(err)
	}

	_, err = h.natsClient.Subscribe("user.deleted", h.deleteUser)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) registerUser(msg *nats.Msg) {
	var userData struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user registered event: %v", err)
		return
	}

	uid := userData.UserID
	email := userData.Email
	firstName := userData.FirstName
	lastName := userData.LastName

	params := generated.InsertUserProfileByIdParams{
		ID:        uuid.MustParse(uid),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	_, err := h.db.GetQueries().InsertUserProfileById(context.Background(), params)
	if err != nil {
		log.Printf("Error creating user profile: %v", err)
		return
	}

	log.Printf("User profile created for user ID: %s", uid)
}

func (h *UserHandler) deleteUser(msg *nats.Msg) {
	var userData struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user deleted event: %v", err)
		return
	}

	uid := userData.UserID

	if err := h.db.GetQueries().DeleteUserProfileById(context.Background(), uuid.MustParse(uid)); err != nil {
		log.Printf("Error deleting user profile: %v", err)
		return
	}

	log.Printf("User profile deleted for user ID: %s", uid)
}
