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
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	calcPb "github.com/mcpxy/mcpx/proto/examples/calculator/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type calculatorServer struct {
	calcPb.UnimplementedCalculatorServiceServer
}

// Add implements the Add method of the CalculatorService.
func (s *calculatorServer) Add(_ context.Context, in *calcPb.AddRequest) (*calcPb.AddResponse, error) {
	slog.Info("grpc_calculator_server: Add called", "a", in.GetA(), "b", in.GetB())
	result := in.GetA() + in.GetB()
	response := &calcPb.AddResponse{}
	response.SetResult(result)
	return response, nil
}

// Subtract implements the Subtract method of the CalculatorService.
func (s *calculatorServer) Subtract(_ context.Context, in *calcPb.SubtractRequest) (*calcPb.SubtractResponse, error) {
	slog.Info("grpc_calculator_server: Subtract called", "a", in.GetA(), "b", in.GetB())
	result := in.GetA() - in.GetB()
	response := &calcPb.SubtractResponse{}
	response.SetResult(result)
	return response, nil
}

// main starts the mock gRPC calculator server.
func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("grpc_calculator_server: Failed to listen", "error", err)
		os.Exit(1)
	}

	actualPort := lis.Addr().(*net.TCPAddr).Port
	slog.Info("grpc_calculator_server: Listening on port", "port", actualPort)
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}

	s := grpc.NewServer()
	calcPb.RegisterCalculatorServiceServer(s, &calculatorServer{})
	reflection.Register(s) // Enable server reflection

	// Graceful shutdown
	go func() {
		if err := s.Serve(lis); err != nil {
			slog.Error("grpc_calculator_server: Failed to serve", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("grpc_calculator_server: Server started.")
	fmt.Println("GRPC_SERVER_READY")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("grpc_calculator_server: Shutting down server...")
	s.GracefulStop()
	slog.Info("grpc_calculator_server: Server shut down gracefully.")
}
