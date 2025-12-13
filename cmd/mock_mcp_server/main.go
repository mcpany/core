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
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Simple MCP server that exposes "read_file" and "list_directory" tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mock-server",
		Version: "1.0.0",
	}, &mcp.ServerOptions{})

	// Register tools
	server.AddTool(&mcp.Tool{
		Name:        "read_file",
		Description: "Read a file",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"path"},
		},
	}, func(_ context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]interface{}
		// request.Params.Arguments is []byte / json.RawMessage
		if err := json.Unmarshal(request.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments json: %w", err)
		}

		path, _ := args["path"].(string)
		if path == "error" {
			return nil, fmt.Errorf("error reading file")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "content of " + path,
				},
			},
		}, nil
	})

	server.AddTool(&mcp.Tool{
		Name:        "list_directory",
		Description: "List directory",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"path"},
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "file1.txt\nfile2.txt",
				},
			},
		}, nil
	})

	// Serve StdIO
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
