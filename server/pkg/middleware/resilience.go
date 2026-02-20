// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"sync"

	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/tool"
)

// ResilienceMiddleware provides circuit breaker and retry functionality for tool executions.
// It wraps tool execution with a resilience manager that handles failures according to policy.
type ResilienceMiddleware struct {
	toolManager tool.ManagerInterface
	managers    sync.Map // map[string]*resilience.Manager (serviceID -> Manager)
}

// NewResilienceMiddleware initializes a new ResilienceMiddleware.
//
// Parameters:
//  toolManager (tool.ManagerInterface): The interface for managing tools.
//
// Returns:
//  *ResilienceMiddleware: The initialized middleware.
func NewResilienceMiddleware(toolManager tool.ManagerInterface) *ResilienceMiddleware {
	return &ResilienceMiddleware{
		toolManager: toolManager,
	}
}

// Execute intercepts tool execution to apply resilience policies.
//
// Parameters:
//  ctx (context.Context): The context for the request.
//  req (*tool.ExecutionRequest): The tool execution request.
//  next (tool.ExecutionFunc): The next handler in the execution chain.
//
// Returns:
//  any: The result of the tool execution.
//  error: An error if the tool execution fails (after retries) or if circuit is open.
//
// Side Effects:
//  Updates circuit breaker state.
//  Retries execution on failure.
func (m *ResilienceMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := m.toolManager.GetTool(req.ToolName)
	if !ok {
		return next(ctx, req)
	}

	serviceID := t.Tool().GetServiceId()
	manager := m.getManager(serviceID)
	if manager == nil {
		return next(ctx, req)
	}

	var result any
	err := manager.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = next(ctx, req)
		return err
	})

	return result, err
}

func (m *ResilienceMiddleware) getManager(serviceID string) *resilience.Manager {
	if val, ok := m.managers.Load(serviceID); ok {
		return val.(*resilience.Manager)
	}

	serviceInfo, ok := m.toolManager.GetServiceInfo(serviceID)
	if !ok || serviceInfo.Config == nil || serviceInfo.Config.GetResilience() == nil {
		// Store nil to avoid repeated lookups if config is missing?
		// But config might be updated later. For now, let's not cache nil eagerly unless we handle updates.
		// However, syncing relies on GetServiceInfo which is fast.
		return nil
	}

	// Double check if config actually has anything enabled
	config := serviceInfo.Config.GetResilience()
	if config.GetCircuitBreaker() == nil && config.GetRetryPolicy() == nil && config.GetTimeout() == nil {
		return nil
	}

	manager := resilience.NewManager(config)

	// We need to use LoadOrStore to avoid race conditions creating multiple managers
	val, loaded := m.managers.LoadOrStore(serviceID, manager)
	if loaded {
		return val.(*resilience.Manager)
	}
	return manager
}
