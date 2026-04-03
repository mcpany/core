// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"sync"

	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/tool"
)

// ResilienceMiddleware represents the public ResilienceMiddleware entity.
//
// Summary: Defines the structured data model representing a middleware.
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
type ResilienceMiddleware struct {
	toolManager tool.ManagerInterface
	managers    sync.Map // map[string]*resilience.Manager (serviceID -> Manager)
}

// NewResilienceMiddleware serves as a public interface for interacting with NewResilienceMiddleware.
//
// Summary: Constructs and returns an initialized resilience middleware ready for consumption.
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
func NewResilienceMiddleware(toolManager tool.ManagerInterface) *ResilienceMiddleware {
	return &ResilienceMiddleware{
		toolManager: toolManager,
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
