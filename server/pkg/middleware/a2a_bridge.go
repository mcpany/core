// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// A2ABridgeMiddleware - Auto-generated documentation.
//
// Summary: A2ABridgeMiddleware represents the Agent-to-Agent (A2A) Bridge middleware.
//
// Fields:
//   - Various fields for A2ABridgeMiddleware.
type A2ABridgeMiddleware struct {
	contextManager *RecursiveContextManager
}

// NewA2ABridgeMiddleware creates a new A2ABridgeMiddleware.
//
// Parameters:
//   - contextManager (*RecursiveContextManager): The manager for A2A session tokens.
//
// Returns:
//   - *A2ABridgeMiddleware: The newly created middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for the middleware struct.
func NewA2ABridgeMiddleware(contextManager *RecursiveContextManager) *A2ABridgeMiddleware {
	return &A2ABridgeMiddleware{
		contextManager: contextManager,
	}
}

// Execute processes the MCP request and intercepts A2A agent calls.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - method (string): The MCP method being called.
//   - req (mcp.Request): The incoming MCP request.
//   - next (mcp.MethodHandler): The next handler in the middleware chain.
//
// Returns:
//   - mcp.Result: The result of the request, either intercepted or from the next handler.
//   - error: Any error that occurred during processing.
//
// Errors:
//   - Returns errors from the next handler if the request is not intercepted.
//
// Side Effects:
//   - May create a new session in the RecursiveContextManager if intercepted.
func (m *A2ABridgeMiddleware) Execute(ctx context.Context, method string, req mcp.Request, next mcp.MethodHandler) (mcp.Result, error) {
	if method != "tools/call" {
		return next(ctx, method, req)
	}

	callReq, ok := req.(*mcp.CallToolRequest)
	if !ok || callReq == nil {
		return next(ctx, method, req)
	}

	if !strings.HasPrefix(callReq.Params.Name, "call_agent_") {
		return next(ctx, method, req)
	}

	// It's an A2A call, intercept it.
	agentName := strings.TrimPrefix(callReq.Params.Name, "call_agent_")

	// Convert arguments to map for session data
	var sessionData map[string]interface{}
	if len(callReq.Params.Arguments) > 0 {
		var mapArgs map[string]interface{}
		if err := json.Unmarshal(callReq.Params.Arguments, &mapArgs); err == nil {
			sessionData = mapArgs
		} else {
			sessionData = map[string]interface{}{"raw_args": string(callReq.Params.Arguments)}
		}
	} else {
		sessionData = map[string]interface{}{}
	}

	// Create a session to store the token for asynchronous callbacks
	session := m.contextManager.CreateSession(sessionData, 1*time.Hour)

	// Return a simulated A2A response
	responseText := fmt.Sprintf("A2A Bridge: Successfully forwarded task to %s. Session ID: %s", agentName, session.ID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
	}, nil
}
