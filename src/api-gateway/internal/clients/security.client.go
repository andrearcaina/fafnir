package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	"fafnir/api-gateway/internal/rbac"
	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/security"

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
	if ok := rbac.IsValidPermission(permission); !ok {
		return &model.HasPermissionResponse{
			Data: &model.SecurityPermission{
				HasPermission: false,
			},
			Code: basepb.ErrorCode_INVALID_ARGUMENT.String(),
		}, nil
	}

	req := &pb.CheckPermissionRequest{
		UserId:     userId,
		Permission: permission,
	}

	resp, err := c.client.CheckPermission(ctx, req)
	if err != nil {
		return &model.HasPermissionResponse{
			Data: &model.SecurityPermission{
				HasPermission: resp.Permission.HasPermission,
			},
			Code: resp.GetCode().String(),
		}, err
	}

	return &model.HasPermissionResponse{
		Data: &model.SecurityPermission{
			HasPermission: resp.Permission.HasPermission,
		},
		Code: resp.Code.String(),
	}, nil
}
