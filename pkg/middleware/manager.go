package middleware

import (
	"context"
	"sync"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager manages the middleware chains for global and service levels.
type Manager struct {
	mu            sync.RWMutex
	globalChain   []Middleware
	serviceChains map[string][]Middleware
}

// NewManager creates a new middleware Manager.
func NewManager() *Manager {
	return &Manager{
		serviceChains: make(map[string][]Middleware),
	}
}

// UpdateConfig updates the middleware configuration.
func (m *Manager) UpdateConfig(globalConfig *configv1.McpAnyServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing middlewares to avoid leaks
	m.closeInternal()

	// Global chain
	var globalChain []Middleware
	if globalConfig.GetGlobalSettings() != nil {
		for _, mwConfig := range globalConfig.GetGlobalSettings().GetMiddleware() {
			mw, err := CreateMiddleware("global", mwConfig)
			if err != nil {
				return err
			}
			if mw != nil {
				globalChain = append(globalChain, mw)
			}
		}

		// Legacy Audit
		if audit := globalConfig.GetGlobalSettings().GetAudit(); audit != nil && audit.GetEnabled() {
			mw, err := NewAuditMiddleware(audit)
			if err == nil && mw != nil {
				globalChain = append(globalChain, mw)
			}
		}
	}
	m.globalChain = globalChain

	// Service chains
	serviceChains := make(map[string][]Middleware)
	for _, svc := range globalConfig.GetUpstreamServices() {
		var chain []Middleware
		for _, mwConfig := range svc.GetMiddleware() {
			mw, err := CreateMiddleware(svc.GetId(), mwConfig)
			if err != nil {
				return err
			}
			if mw != nil {
				chain = append(chain, mw)
			}
		}

		// Compatibility: Support legacy fields
		if svc.GetRateLimit() != nil && svc.GetRateLimit().GetIsEnabled() {
             mw, err := NewRateLimitMiddleware(svc.GetId(), svc.GetRateLimit())
             if err == nil && mw != nil {
                 chain = append(chain, mw)
             }
        }

        if svc.GetCache() != nil && svc.GetCache().GetIsEnabled() {
             mw := NewCachingMiddleware(svc.GetCache())
             chain = append(chain, mw)
        }

        for _, policy := range svc.GetCallPolicies() {
             mw := NewCallPolicyMiddleware(svc.GetId(), policy)
             chain = append(chain, mw)
        }

		serviceChains[svc.GetId()] = chain
		serviceChains[svc.GetName()] = chain
	}
	m.serviceChains = serviceChains

	return nil
}

// Execute executes the middleware chain for a request.
func (m *Manager) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// Identify service
	var serviceID string
	t, ok := tool.GetFromContext(ctx)
	if ok {
		serviceID = t.Tool().GetServiceId()
	}

	m.mu.RLock()
	globalChain := m.globalChain
	serviceChain := m.serviceChains[serviceID]
	m.mu.RUnlock()

	// Combined list: Global -> Service -> Next
	allMiddlewares := make([]Middleware, 0, len(globalChain)+len(serviceChain))
	allMiddlewares = append(allMiddlewares, globalChain...)
	allMiddlewares = append(allMiddlewares, serviceChain...)

	currentHandler := next

	// Iterate backwards to wrap
	for i := len(allMiddlewares) - 1; i >= 0; i-- {
		mw := allMiddlewares[i]
		nextHandler := currentHandler
		currentHandler = func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return mw.Execute(ctx, req, nextHandler)
		}
	}

	return currentHandler(ctx, req)
}

func (m *Manager) closeInternal() {
	closeMiddleware := func(mw Middleware) {
		if c, ok := mw.(interface{ Close() error }); ok {
			_ = c.Close()
		}
	}

	for _, mw := range m.globalChain {
		closeMiddleware(mw)
	}
	for _, chain := range m.serviceChains {
		for _, mw := range chain {
			closeMiddleware(mw)
		}
	}
}

// Close closes all middlewares that implement Closer.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeInternal()
	return nil
}

// ClearCache clears caches. If serviceID is empty, clears all.
func (m *Manager) ClearCache(ctx context.Context, serviceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clearInChain := func(chain []Middleware) {
		for _, mw := range chain {
			if cm, ok := mw.(*CachingMiddleware); ok {
				_ = cm.Clear(ctx)
			}
		}
	}

	if serviceID == "" {
		clearInChain(m.globalChain)
		for _, chain := range m.serviceChains {
			clearInChain(chain)
		}
	} else {
		if chain, ok := m.serviceChains[serviceID]; ok {
			clearInChain(chain)
		}
	}
	return nil
}
