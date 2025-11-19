package api

import (
	"context"
	"fafnir/security-service/internal/db"
	"fafnir/security-service/internal/db/generated"
	basepb "fafnir/shared/pb/base"
	"fafnir/shared/pb/security"

	"github.com/google/uuid"
)

type SecurityHandler struct {
	db *db.Database
	pb.UnimplementedSecurityServiceServer
}

func NewSecurityHandler(database *db.Database) *SecurityHandler {
	return &SecurityHandler{
		db: database,
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
