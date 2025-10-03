/*
 * Copyright 2025 Author(s) of MCPX
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
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/mcpxy/mcpx/proto/examples/calculator/v1"
)

type server struct {
	v1.UnimplementedCalculatorServiceServer
}

func (s *server) Add(ctx context.Context, in *v1.AddRequest) (*v1.AddResponse, error) {
	result := in.GetA() + in.GetB()
	return v1.AddResponse_builder{Result: &result}.Build(), nil
}

func (s *server) Subtract(ctx context.Context, in *v1.SubtractRequest) (*v1.SubtractResponse, error) {
	result := in.GetA() - in.GetB()
	return v1.SubtractResponse_builder{Result: &result}.Build(), nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	v1.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
