package api

import (
	"context"
	"encoding/json"
	"fafnir/security-service/internal/db"
	"fafnir/security-service/internal/db/generated"
	"log"

	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/security"
	natsC "fafnir/shared/pkg/nats"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type SecurityHandler struct {
	db         *db.Database
	natsClient *natsC.NatsClient
	pb.UnimplementedSecurityServiceServer
}

func NewSecurityHandler(database *db.Database, natsClient *natsC.NatsClient) *SecurityHandler {
	return &SecurityHandler{
		db:         database,
		natsClient: natsClient,
	}
}

// CheckPermission implements the gRPC CheckPermission method
func (h *SecurityHandler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.CheckPermissionResponse{
			HasPermission: false,
			Code:          basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	params := generated.CheckUserPermissionParams{
		UserID:         userId,
		PermissionName: req.Permission,
	}

	hasPermission, err := h.db.GetQueries().CheckUserPermission(ctx, params)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return &pb.CheckPermissionResponse{
			HasPermission: false,
			Code:          basepb.ErrorCode_PERMISSION_DENIED,
		}, nil
	}

	return &pb.CheckPermissionResponse{
		HasPermission: true,
		Code:          basepb.ErrorCode_OK,
	}, nil
}

func (h *SecurityHandler) RegisterSubscribeHandlers() {
	_, err := h.natsClient.Subscribe("user.registered", h.registerUser)
	if err != nil {
		log.Fatal(err)
	}

	_, err = h.natsClient.Subscribe("user.deleted", h.deleteUser)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *SecurityHandler) registerUser(msg *nats.Msg) {
	var userData struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user registered event: %v", err)
		return
	}

	uid := userData.UserID

	params := generated.InsertUserRoleWithIDParams{
		UserID:   uuid.MustParse(uid),
		RoleName: "member", // hardcoded default for new users
	}

	_, err := h.db.GetQueries().InsertUserRoleWithID(context.Background(), params)
	if err != nil {
		log.Printf("Error creating user profile: %v", err)
		return
	}

	log.Printf("User profile created for user ID: %s", uid)
}

func (h *SecurityHandler) deleteUser(msg *nats.Msg) {
	var userData struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user deleted event: %v", err)
		return
	}

	uid := userData.UserID

	err := h.db.GetQueries().DeleteUserRoleWithID(context.Background(), uuid.MustParse(uid))
	if err != nil {
		log.Printf("Error deleting user roles: %v", err)
		return
	}

	log.Printf("User roles deleted for user ID: %s", uid)
}
