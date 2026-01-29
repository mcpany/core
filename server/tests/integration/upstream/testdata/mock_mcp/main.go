// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	// JSONRPC is the protocol version (must be "2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the request identifier.
	ID *json.RawMessage `json:"id,omitempty"`
	// Method is the method name.
	Method string `json:"method"`
	// Params is the parameter object or array.
	Params json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	// JSONRPC is the protocol version (must be "2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the request identifier.
	ID *json.RawMessage `json:"id,omitempty"`
	// Result is the successful result.
	Result interface{} `json:"result,omitempty"`
	// Error is the error object.
	Error *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`
	// Message is the error message.
	Message string `json:"message"`
}

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	log.Println("Mock MCP Server Started")

	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				return
			}
			log.Printf("Error decoding: %v", err)
			continue
		}

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		switch req.Method {
		case "initialize":
			resp.Result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "mock-mcp",
					"version": "1.0.0",
				},
			}
		case "notifications/initialized":
			// No response needed
			continue
		case "tools/list", "mcp.listTools":
			resp.Result = map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "get_html",
						"description": "Returns a simple HTML string",
						"inputSchema": map[string]interface{}{
							"type": "object",
						},
					},
				},
			}
		case "tools/call", "mcp.callTool":
			// Parse params to check name if needed, but we only have one tool
			resp.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "<h1>Mock Title</h1><p>Mock content</p>",
					},
				},
			}
		case "ping":
			resp.Result = map[string]interface{}{}
		case "prompts/list", "mcp.listPrompts":
			resp.Result = map[string]interface{}{
				"prompts": []interface{}{},
			}
		case "resources/list", "mcp.listResources":
			resp.Result = map[string]interface{}{
				"resources": []interface{}{},
			}
		default:
			// Ignore or error
			// For initialization flow, valid requests are enough
		}

		if req.ID != nil {
			if err := encoder.Encode(resp); err != nil {
				log.Printf("Error encoding: %v", err)
			}
		}
	}
}
