// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
		log.Fatalf("Failed to connect to MCPANY server: %v", err) // nolint:gocritic
	}
	defer func() { _ = cs.Close() }()

	result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
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
