package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	basepb "fafnir/shared/pb/base"
	"fafnir/shared/pb/security"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SecurityClient struct {
	conn   *grpc.ClientConn
	client pb.SecurityServiceClient
}

func NewSecurityClient(address string) *SecurityClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil
	}

	client := pb.NewSecurityServiceClient(conn)

	return &SecurityClient{
		conn:   conn,
		client: client,
	}
}

func (c *SecurityClient) CheckPermission(ctx context.Context, userId string, permission string) (*model.HasPermissionResponse, error) {
	req := &pb.CheckPermissionRequest{
		UserId:     userId,
		Permission: permission,
	}

	resp, err := c.client.CheckPermission(ctx, req)
	if err != nil {
		return &model.HasPermissionResponse{
			HasPermission:  false,
			PermissionCode: basepb.ErrorCode_PERMISSION_DENIED.String(),
		}, err
	}

	return &model.HasPermissionResponse{
		HasPermission:  resp.HasPermission,
		PermissionCode: resp.Code.String(),
	}, nil
}
