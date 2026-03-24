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

// Summary: A2ABridgeMiddleware represents the Agent-to-Agent (A2A) Bridge middleware. It intercepts tool calls prefixed with "call_agent_" and bridges them to the A2A protocol. Represents a A2ABridgeMiddleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type A2ABridgeMiddleware struct {
	contextManager *RecursiveContextManager
}

// Summary: NewA2ABridgeMiddleware creates a new A2ABridgeMiddleware.
//
// Parameters:
//   - contextManager (*RecursiveContextManager): The contextManager parameter.
//
// Returns:
//   - *A2ABridgeMiddleware: The resulting *A2ABridgeMiddleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewA2ABridgeMiddleware(contextManager *RecursiveContextManager) *A2ABridgeMiddleware {
	return &A2ABridgeMiddleware{
		contextManager: contextManager,
	}
}

// Summary: Execute processes the MCP request and intercepts A2A agent calls.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - method (string): The method parameter.
//   - req (mcp.Request): The req parameter.
//   - next (mcp.MethodHandler): The next parameter.
//
// Returns:
//   - mcp.Result: The resulting mcp.Result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
