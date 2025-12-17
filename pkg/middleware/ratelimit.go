// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"sync"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality for upstream services.
type RateLimitMiddleware struct {
	toolManager tool.ManagerInterface
	limiters    sync.Map // map[string]*rate.Limiter
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		toolManager: toolManager,
	}
}

// Execute executes the rate limiting middleware.
func (m *RateLimitMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := tool.GetFromContext(ctx)
	if !ok {
		return next(ctx, req)
	}

	serviceID := t.Tool().GetServiceId()
	serviceInfo, ok := m.toolManager.GetServiceInfo(serviceID)
	if !ok {
		// If service info is not found, we cannot enforce rate limits.
		// Proceed with execution.
		return next(ctx, req)
	}

	rateLimitConfig := serviceInfo.Config.GetRateLimit()
	if rateLimitConfig == nil || !rateLimitConfig.GetIsEnabled() {
		return next(ctx, req)
	}

	limiter := m.getLimiter(serviceID, rateLimitConfig)
	if !limiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded for service %s", serviceInfo.Name)
	}

	return next(ctx, req)
}

func (m *RateLimitMiddleware) getLimiter(serviceID string, config *configv1.RateLimitConfig) *rate.Limiter {
	rps := rate.Limit(config.GetRequestsPerSecond())
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	limiterVal, ok := m.limiters.Load(serviceID)
	if !ok {
		newLimiter := rate.NewLimiter(rps, burst)
		actual, loaded := m.limiters.LoadOrStore(serviceID, newLimiter)
		limiterVal = actual
		// If we loaded an existing one (race condition), we still fall through to check/update below.
		// This handles the case where the race winner had an older config (unlikely but possible).
		_ = loaded
	}

	limiter := limiterVal.(*rate.Limiter)

	// Check and update if config has changed.
	// Note: SetLimit and SetBurst are thread-safe.
	if limiter.Limit() != rps {
		limiter.SetLimit(rps)
	}
	if limiter.Burst() != burst {
		limiter.SetBurst(burst)
	}

	return limiter
}
