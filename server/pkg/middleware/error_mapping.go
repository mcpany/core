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

// ErrorMappingMiddleware normalizes diverse upstream errors into standard MCP errors.
//
// Summary: Normalizes arbitrary tool execution errors.
type ErrorMappingMiddleware struct{}

// NewErrorMappingMiddleware creates a new error mapping middleware.
//
// Summary: Initializes the middleware responsible for translating internal errors into safe external responses.
//
// Parameters:
//   - None.
//
// Returns:
//   - mcp.Middleware: The initialized error mapping middleware.
//
// Throws/Errors:
//   - None.
func NewErrorMappingMiddleware() *ErrorMappingMiddleware {
	return &ErrorMappingMiddleware{}
}

// Execute performs the middleware logic, wrapping the next handler.
//
// Summary: Executes middleware logic.
//
// Parameters:
//   - ctx (context.Context): The context for the execution.
//   - req (*tool.ExecutionRequest): The execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if execution fails.
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
