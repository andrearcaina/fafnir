package api

import (
	"context"
	basepb "fafnir/shared/pb/base"
	"fafnir/shared/pb/user"
	"fafnir/user-service/internal/db"

	"github.com/google/uuid"
)

type UserHandler struct {
	db *db.Database
	pb.UnimplementedUserServiceServer
}

func NewUserHandler(database *db.Database) *UserHandler {
	return &UserHandler{
		db: database,
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
