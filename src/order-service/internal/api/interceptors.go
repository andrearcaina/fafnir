package api

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// log request
	log.Printf("gRPC Request - Method: %s", info.FullMethod)
	log.Printf("gRPC Request - Payload: %+v", req)

	// call the handler
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("gRPC Error - Method: %s, Duration: %v, Error: %v", info.FullMethod, duration, err)
		return nil, err
	}

	// log response (no need for different log handling based on method, unless needed later)
	log.Printf("gRPC Response - Method: %s, Duration: %v, Response: %+v", info.FullMethod, duration, resp)

	return resp, nil
}
