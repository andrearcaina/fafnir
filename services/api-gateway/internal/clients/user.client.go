package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	basepb "fafnir/shared/pb/base"
	"fafnir/shared/pb/user"

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
		return &model.ProfileDataResponse{
			UserID:         "",
			FirstName:      "",
			LastName:       "",
			PermissionCode: basepb.ErrorCode_NOT_FOUND.String(),
		}, err
	}

	return &model.ProfileDataResponse{
		UserID:         resp.UserId,
		FirstName:      resp.FirstName,
		LastName:       resp.LastName,
		PermissionCode: resp.Code.String(),
	}, nil
}
