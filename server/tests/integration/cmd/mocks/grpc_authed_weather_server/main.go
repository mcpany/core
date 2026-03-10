// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock gRPC authenticated weather server for testing.
package main

import (
	"context" //nolint:gci
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	weatherV1 "github.com/mcpany/core/proto/examples/weather/v1"
)

const (
	expectedAuthScheme = "bearer"
	expectedAuthToken  = "test-bearer-token"
)

type server struct {
	weatherV1.UnimplementedWeatherServiceServer
}

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

// GetWeather returns the weather for a specific location.
//
// Parameters:
//   ctx: The context for the request.
//   req: The request containing the location.
//
// Returns:
//   *weatherV1.GetWeatherResponse: The response containing the weather description.
//   error: An error if the location is not found.
func (s *server) GetWeather(_ context.Context, req *weatherV1.GetWeatherRequest) (*weatherV1.GetWeatherResponse, error) {
	log.Printf("INFO grpc_authed_weather_server: GetWeather called location=%s", req.GetLocation())
	weather, ok := weatherData[req.GetLocation()]
	if !ok {
		return nil, fmt.Errorf("location not found")
	}
	response := &weatherV1.GetWeatherResponse{}
	response.SetWeather(weather)
	return response, nil
}

func authInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

	var lis net.Listener
	var err error
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	for i := 0; i < 5; i++ {
		var lc net.ListenConfig
		lis, err = lc.Listen(context.Background(), "tcp", addr)
		if err == nil {
			break
		}
		log.Printf("WARN grpc_authed_weather_server: Failed to listen, retrying... error=%v attempt=%d", err, i+1)
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("failed to listen after retries: %v", err)
	}

	actualPort := lis.Addr().(*net.TCPAddr).Port
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
		_ = os.Stdout.Sync()
	}

	s := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(authInterceptor),
	)
	weatherV1.RegisterWeatherServiceServer(s, &server{})
	reflection.Register(s)

	log.Printf("INFO grpc_authed_weather_server: Listening on port port=%d", actualPort)
	fmt.Println("GRPC_SERVER_READY") // Signal that the server is ready
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
