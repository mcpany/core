// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock prompt server for integration tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	port := flag.Int("port", 0, "The server port")
	flag.Parse()

	// 1. Create a new MCP Server
	serverImpl := &mcp.Implementation{Name: "e2e-prompt-server", Version: "v0.0.1"}
	serverOpts := &mcp.ServerOptions{HasPrompts: true}
	server := mcp.NewServer(serverImpl, serverOpts)

	// 2. Add a prompt
	prompt := &mcp.Prompt{
		Name:        "hello",
		Description: "A simple hello world prompt",
	}
	server.AddPrompt(prompt, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Hello, world!"},
				},
			},
		}, nil
	})

	// 3. Create a streamable HTTP handler
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	// 4. Serve the handler over HTTP
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Log the actual address (including dynamic port)
	log.Printf("server listening at %s", listener.Addr().String())

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := srv.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
