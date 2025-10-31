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
	"flag"
	"fmt"
	"log"
	"net/http"

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
	server.AddPrompt(prompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	// 4. Serve the handler over HTTP
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("server listening at %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
