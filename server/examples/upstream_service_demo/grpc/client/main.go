// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/mcpany/core/upstream_service/grpc/greeter_server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func run() error {
	// Set up a connection to the server.
	// Use NewClient instead of Dial to resolve lint warning SA1019.
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		return err
	}
	log.Printf("Greeting: %s", r.GetMessage())
	return nil
}
