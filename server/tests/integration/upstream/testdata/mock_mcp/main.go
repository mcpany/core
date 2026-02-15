package main

import (
	"encoding/json"
	"log"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int    `json:"code"`
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
