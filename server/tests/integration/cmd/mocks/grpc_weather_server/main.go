// Package main implements a mock gRPC weather server.

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

	weatherPb "github.com/mcpany/core/proto/examples/weather/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type weatherServer struct {
	weatherPb.UnimplementedWeatherServiceServer
}

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

// GetWeather implements the GetWeather method of the WeatherService.
func (s *weatherServer) GetWeather(_ context.Context, in *weatherPb.GetWeatherRequest) (*weatherPb.GetWeatherResponse, error) {
	slog.Info("grpc_weather_server: GetWeather called", "location", in.GetLocation())
	weather, ok := weatherData[in.GetLocation()]
	if !ok {
		return nil, fmt.Errorf("location not found")
	}
	response := &weatherPb.GetWeatherResponse{}
	response.SetWeather(weather)
	return response, nil
}

// main starts the mock gRPC weather server.
func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("grpc_weather_server: Failed to listen", "error", err)
		os.Exit(1)
	}

	actualPort := lis.Addr().(*net.TCPAddr).Port
	slog.Info("grpc_weather_server: Listening on port", "port", actualPort)
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}

	s := grpc.NewServer()
	weatherPb.RegisterWeatherServiceServer(s, &weatherServer{})
	reflection.Register(s) // Enable server reflection

	// Graceful shutdown
	go func() {
		if err := s.Serve(lis); err != nil {
			slog.Error("grpc_weather_server: Failed to serve", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("grpc_weather_server: Server started.")
	fmt.Println("GRPC_SERVER_READY")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("grpc_weather_server: Shutting down server...")
	s.GracefulStop()
	slog.Info("grpc_weather_server: Server shut down gracefully.")
}
