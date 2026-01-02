package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/user"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserClient(address string) *UserClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil
	}

	client := pb.NewUserServiceClient(conn)

	return &UserClient{
		conn:   conn,
		client: client,
	}
}

func (c *UserClient) GetProfileData(ctx context.Context, userId string) (*model.ProfileDataResponse, error) {
	req := &pb.ProfileDataRequest{
		UserId: userId,
	}

	resp, err := c.client.GetProfileData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile data: %w", err)
	}

	if resp.GetCode() != basepb.ErrorCode_OK {
		return &model.ProfileDataResponse{
			Data: nil,
			Code: resp.GetCode().String(),
		}, nil
	}

	return &model.ProfileDataResponse{
		Data: &model.ProfileData{
			UserID:    resp.Data.UserId,
			FirstName: resp.Data.FirstName,
			LastName:  resp.Data.LastName,
		},
		Code: resp.GetCode().String(),
	}, nil
}
