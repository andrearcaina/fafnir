package api

import (
	"context"
	pb "fafnir/shared/pb/user"
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
	if info.FullMethod == "/user.UserService/GetProfileData" {
		logGetProfileData(duration, resp.(*pb.ProfileDataResponse))
	}

	return resp, nil
}

func logGetProfileData(duration time.Duration, resp *pb.ProfileDataResponse) {
	log.Printf("gRPC Response - Method: /user.UserService/GetProfileData, Duration: %v, UserId: %s, FirstName: %s, LastName: %s, Code: %s",
		duration,
		resp.GetUserId(),
		resp.GetFirstName(),
		resp.GetLastName(),
		resp.GetCode().String(),
	)
}
