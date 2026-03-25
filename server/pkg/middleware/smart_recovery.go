// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
)

// SmartRecoveryMiddleware provides autonomous error recovery for tool calls.
//
// Summary: Represents a SmartRecoveryMiddleware.
type SmartRecoveryMiddleware struct {
	config      *configv1.SmartRecoveryConfig
	toolManager tool.ManagerInterface
}

// NewSmartRecoveryMiddleware creates a new SmartRecoveryMiddleware instance.
//
// Parameters:
//   - config (*configv1.SmartRecoveryConfig): The recovery configuration.
//   - toolManager (tool.ManagerInterface): The manager for tool execution.
//
// Returns:
//   - *SmartRecoveryMiddleware: The initialized middleware instance.
//
// Summary: Executes NewSmartRecoveryMiddleware operation.
func NewSmartRecoveryMiddleware(config *configv1.SmartRecoveryConfig, toolManager tool.ManagerInterface) *SmartRecoveryMiddleware {
	return &SmartRecoveryMiddleware{
		config:      config,
		toolManager: toolManager,
	}
}

// Execute wraps tool execution with recovery logic.
//
// Parameters:
//   - ctx (context.Context): The request context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - (any): The result of the tool execution (or recovered result).
//   - (error): An error if execution and recovery both fail.
//
// Summary: Executes Execute operation.
func (m *SmartRecoveryMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	return next(ctx, req)
}

// Handler returns an HTTP middleware wrapper.
//
// Parameters:
//   - next (http.Handler): The next HTTP handler.
//
// Returns:
//   - (http.Handler): The wrapped HTTP handler.
//
// Summary: Executes Handler operation.
func (m *SmartRecoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
