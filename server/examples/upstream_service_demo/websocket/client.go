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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: "http://localhost:8081"}, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCPANY server: %w", err)
	}
	defer func() { _ = cs.Close() }()

	result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	for _, tool := range result.Tools {
		fmt.Printf("Tool: %s\n", tool.Name)
	}

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "echo-service.echo", Arguments: map[string]interface{}{"message": "hello"}})
	if err != nil {
		return fmt.Errorf("error calling tool: %w", err)
	}

	if res.IsError {
		return fmt.Errorf("tool returned an error: %v", res.Content)
	}

	if len(res.Content) == 0 {
		return fmt.Errorf("expected content in tool response")
	}

	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		return fmt.Errorf("expected content to be of type TextContent")
	}

	var toolResult map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &toolResult); err != nil {
		return fmt.Errorf("failed to unmarshal tool output: %w", err)
	}

	fmt.Printf("Tool result: %v\n", toolResult)
	return nil
}
