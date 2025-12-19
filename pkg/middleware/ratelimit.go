// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"sync"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"golang.org/x/time/rate"
)

// Limiter interface defines the methods required for a rate limiter.
type Limiter interface {
	Allow(ctx context.Context) (bool, error)
	Update(rps float64, burst int)
}

// LocalLimiter is an in-memory implementation of Limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed.
func (l *LocalLimiter) Allow(ctx context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// Update updates the limiter configuration.
func (l *LocalLimiter) Update(rps float64, burst int) {
	limit := rate.Limit(rps)
	if l.Limiter.Limit() != limit {
		l.Limiter.SetLimit(limit)
	}
	if l.Limiter.Burst() != burst {
		l.Limiter.SetBurst(burst)
	}
}

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality for upstream services.
type RateLimitMiddleware struct {
	toolManager tool.ManagerInterface
	limiters    sync.Map // map[string]Limiter
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		toolManager: toolManager,
	}
}

// Execute executes the rate limiting middleware.
func (m *RateLimitMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := m.toolManager.GetTool(req.ToolName)
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

	limiter, err := m.getLimiter(serviceID, rateLimitConfig)
	if err != nil {
		// Failed to get limiter (e.g. redis config error).
		// Fail open.
		return next(ctx, req)
	}

	allowed, err := limiter.Allow(ctx)
	if err != nil {
		// If check fails (e.g. redis connection error).
		// Fail open.
		return next(ctx, req)
	}

	if !allowed {
		metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
			{Name: "service_id", Value: serviceID},
			{Name: "status", Value: "blocked"},
		})
		return nil, fmt.Errorf("rate limit exceeded for service %s", serviceInfo.Name)
	}

	metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
		{Name: "service_id", Value: serviceID},
		{Name: "status", Value: "allowed"},
	})

	return next(ctx, req)
}

func (m *RateLimitMiddleware) getLimiter(serviceID string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	limiterVal, ok := m.limiters.Load(serviceID)
	var limiter Limiter
	isRedis := config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS

	if ok {
		limiter = limiterVal.(Limiter)
		// Check if type matches config
		if isRedis {
			if _, ok := limiter.(*RedisLimiter); !ok {
				limiter = nil // Type mismatch, force recreate
			}
		} else {
			if _, ok := limiter.(*LocalLimiter); !ok {
				limiter = nil // Type mismatch, force recreate
			}
		}
	}

	if limiter == nil {
		if isRedis {
			redisLimiter, err := NewRedisLimiter(serviceID, config)
			if err != nil {
				return nil, err
			}
			limiter = redisLimiter
		} else {
			limiter = &LocalLimiter{
				Limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
		}
		m.limiters.Store(serviceID, limiter)
	} else {
		// Update existing
		limiter.Update(rps, burst)
	}

	return limiter, nil
}
