/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	calculatorv1 "github.com/mcpxy/core/proto/examples/calculator/v1"
)

const (
	expectedAuthScheme = "bearer"
	expectedAuthToken  = "test-bearer-token"
)

type server struct {
	calculatorv1.UnimplementedCalculatorServiceServer
}

func (s *server) Add(ctx context.Context, req *calculatorv1.AddRequest) (*calculatorv1.AddResponse, error) {
	log.Printf("INFO grpc_authed_calculator_server: Add called a=%d b=%d", req.GetA(), req.GetB())
	result := req.GetA() + req.GetB()
	return calculatorv1.AddResponse_builder{
		Result: &result,
	}.Build(), nil
}

func (s *server) Subtract(ctx context.Context, req *calculatorv1.SubtractRequest) (*calculatorv1.SubtractResponse, error) {
	log.Printf("INFO grpc_authed_calculator_server: Subtract called a=%d b=%d", req.GetA(), req.GetB())
	result := req.GetA() - req.GetB()
	return calculatorv1.SubtractResponse_builder{
		Result: &result,
	}.Build(), nil
}

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	authHeader := authHeaders[0]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != expectedAuthScheme {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	if parts[1] != expectedAuthToken {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return handler(ctx, req)
}

func main() {
	port := flag.Int("port", 0, "Port to listen on")
	flag.Parse()

	if *port == 0 {
		log.Fatal("port is required")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(authInterceptor),
	)
	calculatorv1.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)

	log.Printf("INFO grpc_authed_calculator_server: Listening on port port=%d", *port)
	fmt.Println("GRPC_SERVER_READY") // Signal that the server is ready
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
