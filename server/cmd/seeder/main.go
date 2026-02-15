package main

import (
	"context"
	"log"
	"os"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func main() {
	addr := os.Getenv("MCPANY_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	log.Printf("Connecting to %s...", addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := apiv1.NewRegistrationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create a dummy service using Builders
	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("seed_tool"),
		Description: proto.String("A seeded tool for testing"),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String("http://localhost:8080"),
		Tools:   []*configv1.ToolDefinition{toolDef},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("seed_service"),
		Version:     proto.String("v1.0.0"),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	resp, err := c.RegisterService(ctx, req)
	if err != nil {
		log.Fatalf("could not register service: %v", err)
	}

	log.Printf("Service registered: %s", resp.GetMessage())
}
