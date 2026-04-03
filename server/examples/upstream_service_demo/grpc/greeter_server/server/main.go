// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/mcpany/core/upstream_service/grpc/greeter_server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement greeter.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello serves as a public interface for interacting with SayHello.
//
// Summary: Say the hello appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	addr := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
