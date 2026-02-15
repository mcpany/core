package middleware

import (
	"context"
	"sync"

	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/tool"
)

// ResilienceMiddleware provides circuit breaker and retry functionality for tool executions.
type ResilienceMiddleware struct {
	toolManager tool.ManagerInterface
	managers    sync.Map // map[string]*resilience.Manager (serviceID -> Manager)
}

// NewResilienceMiddleware creates a new ResilienceMiddleware.
//
// toolManager is the toolManager.
//
// Returns the result.
func NewResilienceMiddleware(toolManager tool.ManagerInterface) *ResilienceMiddleware {
	return &ResilienceMiddleware{
		toolManager: toolManager,
	}
}

// Execute executes the resilience middleware.
//
// ctx is the context for the request.
// req is the request object.
// next is the next.
//
// Returns the result.
// Returns an error if the operation fails.
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
