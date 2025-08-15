package api

import (
	"context"
	"fafnir/security-service/internal/db"
	"fafnir/security-service/internal/db/generated"
	"fafnir/shared/pb/security"

	"github.com/google/uuid"
)

type Handler struct {
	db *db.Database
	pb.UnimplementedSecurityServiceServer
}

func NewSecurityHandler(database *db.Database) *Handler {
	return &Handler{
		db: database,
	}
}

// CheckPermission implements the gRPC CheckPermission method
func (h *Handler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.CheckPermissionResponse{
			HasPermission: false,
			Code:          pb.ErrorCode_INVALID_ARGUMENT,
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
			Code:          pb.ErrorCode_PERMISSION_DENIED,
		}, nil
	}

	return &pb.CheckPermissionResponse{
		HasPermission: true,
		Code:          pb.ErrorCode_OK,
	}, nil
}
