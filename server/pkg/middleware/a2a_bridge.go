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

// A2ABridgeMiddleware represents the public A2ABridgeMiddleware entity.
//
// Summary: Defines the structured data model representing a bridge middleware.
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

// NewA2ABridgeMiddleware serves as a public interface for interacting with NewA2ABridgeMiddleware.
//
// Summary: Constructs and returns an initialized a bridge middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewA2ABridgeMiddleware(contextManager *RecursiveContextManager) *A2ABridgeMiddleware {
	return &A2ABridgeMiddleware{
		contextManager: contextManager,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
