// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a demo WebSocket client.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: "http://localhost:8081"}, nil)
	if err != nil {
		cancel()
		log.Printf("Failed to connect to MCPANY server: %v", err)
		return
	}
	defer func() { _ = cs.Close() }()

	result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err) //nolint:gocritic // Example code, exit is intended
	}

	for _, tool := range result.Tools {
		fmt.Printf("Tool: %s\n", tool.Name)
	}

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "echo-service.echo", Arguments: map[string]interface{}{"message": "hello"}})
	if err != nil {
		log.Fatalf("Error calling tool: %v", err)
	}

	if res.IsError {
		log.Fatalf("Tool returned an error: %v", res.Content)
	}

	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		log.Fatalf("Expected content to be of type TextContent")
	}

	var toolResult map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &toolResult); err != nil {
		log.Fatalf("Failed to unmarshal tool output: %v", err)
	}

	fmt.Printf("Tool result: %v\n", toolResult)
}
