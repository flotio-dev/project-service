package main

import (
	"context"
	"log"
	"time"

	grpcpb "github.com/flotio-dev/user-service/internal/userpb"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := grpcpb.NewUserServiceClient(conn)
	_ = client

	// Example usage: set a timeout context for requests
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// _ = client // Use client for RPC calls
	_ = ctx // Remove or use ctx as needed

	// Add your client calls here
}
