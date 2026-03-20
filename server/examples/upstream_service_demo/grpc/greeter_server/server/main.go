// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements the main subsystem.
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

// SayHello handles say hello.
//
// Parameters:
//   - ctx (context.Context): The execution context for managing deadlines and cancellations.
//   - in (*pb.HelloRequest): The in input required for processing.
//
// Returns:
//   - *pb.HelloReply: The generated or retrieved entity.
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// SayHello implements greeter.GreeterServer
//
// ctx is the context for the request.
// in is the request object.
//
// Returns the result.
// Returns an error if the operation fails.
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
