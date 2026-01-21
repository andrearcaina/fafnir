package api

import (
	"context"
	"encoding/json"
	"fafnir/security-service/internal/db"
	"fafnir/security-service/internal/db/generated"
	"fmt"
	"log"

	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/security"
	natsC "fafnir/shared/pkg/nats"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	lru "github.com/hashicorp/golang-lru/v2"
)

type SecurityHandler struct {
	db         *db.Database
	natsClient *natsC.NatsClient
	pb.UnimplementedSecurityServiceServer

	// key: "userID:permission", value: true/false
	// using an LRU cache to limit memory usage, we don't need this to be in redis as it's not critical data
	// it is just for super fast permission checks within the service
	permissionCache *lru.Cache[string, bool]
}

func NewSecurityHandler(database *db.Database, natsClient *natsC.NatsClient) *SecurityHandler {
	cache, err := lru.New[string, bool](1000) // Cache size of 1000 entries
	if err != nil {
		log.Fatalf("Failed to create permission cache: %v", err)
	}

	return &SecurityHandler{
		db:              database,
		natsClient:      natsClient,
		permissionCache: cache,
	}
}

// CheckPermission implements the gRPC CheckPermission method
func (h *SecurityHandler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	// check cache first
	cachedKey := fmt.Sprintf("%s:%s", req.UserId, req.Permission)

	// if found in cache, return cached value
	if hasPermission, found := h.permissionCache.Get(cachedKey); found {
		log.Printf("Permission cache hit for key: %s", cachedKey)
		return &pb.CheckPermissionResponse{
			Permission: &pb.SecurityPermission{
				HasPermission: hasPermission,
			},
			Code: basepb.ErrorCode_OK,
		}, nil
	}

	// else, check database
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.CheckPermissionResponse{
			Permission: &pb.SecurityPermission{
				HasPermission: false,
			},
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
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
			Permission: &pb.SecurityPermission{
				HasPermission: hasPermission,
			},
			Code: basepb.ErrorCode_PERMISSION_DENIED,
		}, nil
	}

	// cache the positive result for future requests from the same user
	h.permissionCache.Add(cachedKey, true)

	return &pb.CheckPermissionResponse{
		Permission: &pb.SecurityPermission{
			HasPermission: hasPermission,
		},
		Code: basepb.ErrorCode_OK,
	}, nil
}

func (h *SecurityHandler) RegisterSubscribeHandlers() {
	_, err := h.natsClient.QueueSubscribe("users.>", "security-service-main", "security-users-consumer", h.handleUserEvents)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *SecurityHandler) handleUserEvents(msg *nats.Msg) {
	var err error

	switch msg.Subject {
	case "users.registered":
		err = h.registerUser(msg)
	case "users.deleted":
		err = h.deleteUser(msg)
	default:
		// ignore events we don't care about
		// we must ack them, otherwise they come back forever
		_ = msg.Ack()
		return
	}

	if err != nil {
		log.Printf("Failed to process %s: %v", msg.Subject, err)
		_ = msg.Nak() // retry later (negative ack)
	} else {
		_ = msg.Ack() // success (acknowledge message)
	}
}

func (h *SecurityHandler) registerUser(msg *nats.Msg) error {
	var userData struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user registered event: %v", err)
		return nil
	}

	uid := userData.UserID

	params := generated.InsertUserRoleWithIDParams{
		UserID:   uuid.MustParse(uid),
		RoleName: "member", // hardcoded default for new users
	}

	_, err := h.db.GetQueries().InsertUserRoleWithID(context.Background(), params)
	if err != nil {
		log.Printf("Error creating user profile: %v", err)
		return err
	}

	log.Printf("User profile created for user ID: %s", uid)
	return nil
}

func (h *SecurityHandler) deleteUser(msg *nats.Msg) error {
	var userData struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Data, &userData); err != nil {
		log.Printf("Error unmarshaling user deleted event: %v", err)
		return nil
	}

	uid := userData.UserID

	err := h.db.GetQueries().DeleteUserRoleWithID(context.Background(), uuid.MustParse(uid))
	if err != nil {
		log.Printf("Error deleting user roles: %v", err)
		return err
	}

	log.Printf("User roles deleted for user ID: %s", uid)
	return nil
}
