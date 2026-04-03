// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrorMappingMiddleware represents the public ErrorMappingMiddleware entity.
//
// Summary: Defines the structured data model representing a mapping middleware.
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
type ErrorMappingMiddleware struct{}

// NewErrorMappingMiddleware serves as a public interface for interacting with NewErrorMappingMiddleware.
//
// Summary: Constructs and returns an initialized error mapping middleware ready for consumption.
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
func NewErrorMappingMiddleware() *ErrorMappingMiddleware {
	return &ErrorMappingMiddleware{}
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
func (m *ErrorMappingMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	res, err := next(ctx, req)
	if err != nil {
		// Log the original error for debugging purposes
		logging.GetLogger().DebugContext(ctx, "ErrorMappingMiddleware caught error", "tool", req.ToolName, "error", err)

		// Map to standard MCP CallToolResult with IsError=true
		return mapToMCPErrorResult(err), nil
	}
	return res, nil
}

// mapToMCPErrorResult takes a generic error and attempts to map it to a standard mcp.CallToolResult with IsError=true.
//
// Summary: Maps an error to mcp.CallToolResult.
//
// Parameters:
//   - err (error): The original error.
//
// Returns:
//   - *mcp.CallToolResult: The standardized MCP error result.
func mapToMCPErrorResult(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Heuristics to detect common error types and map them to appropriate JSON-RPC/MCP error messages.
	// We use IsError=true to signal to the client that the tool call failed,
	// while providing a standardized error message in the content block.

	var message string
	switch {
	case strings.Contains(strings.ToLower(errStr), "not found") || strings.Contains(strings.ToLower(errStr), "no such file"):
		message = fmt.Sprintf("Resource not found: %v", err)
	case strings.Contains(strings.ToLower(errStr), "permission denied") || strings.Contains(strings.ToLower(errStr), "forbidden"):
		message = fmt.Sprintf("Permission denied: %v", err)
	case strings.Contains(strings.ToLower(errStr), "timeout") || strings.Contains(strings.ToLower(errStr), "deadline exceeded"):
		message = fmt.Sprintf("Upstream timeout: %v", err)
	case strings.Contains(strings.ToLower(errStr), "invalid syntax") || strings.Contains(strings.ToLower(errStr), "bad request"):
		message = fmt.Sprintf("Invalid parameters: %v", err)
	case strings.Contains(strings.ToLower(errStr), "unauthorized"):
		message = fmt.Sprintf("Unauthorized: %v", err)
	default:
		message = fmt.Sprintf("Upstream execution error: %v", err)
	}

	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: message,
			},
		},
	}
}
