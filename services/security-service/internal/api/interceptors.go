package api

import (
	"context"
	pb "fafnir/shared/pb/security"
	"log"
	"time"

	"google.golang.org/grpc"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Log request
	log.Printf("gRPC Request - Method: %s", info.FullMethod)
	log.Printf("gRPC Request - Payload: %+v", req)

	// Call the handler
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("gRPC Error - Method: %s, Duration: %v, Error: %v", info.FullMethod, duration, err)
		return nil, err
	}

	// Log response
	if info.FullMethod == "/security.SecurityService/CheckPermission" {
		logCheckPermission(duration, resp)
	}

	return resp, nil
}

func logCheckPermission(duration time.Duration, resp interface{}) {
	log.Printf(
		"gRPC Response - Duration: %v, Response: has_permission=%v, code=%s",
		duration,
		resp.(*pb.CheckPermissionResponse).GetHasPermission(),
		resp.(*pb.CheckPermissionResponse).GetCode().String(),
	)
}
