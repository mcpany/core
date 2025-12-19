// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"golang.org/x/time/rate"
)

// Limiter interface defines the methods required for a rate limiter.
type Limiter interface {
	// Allow checks if the request is allowed.
	Allow(ctx context.Context) (bool, error)
	// Update updates the limiter configuration.
	Update(rps float64, burst int)
	// Close cleans up resources.
	Close() error
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

// Close is a no-op for LocalLimiter.
func (l *LocalLimiter) Close() error {
	return nil
}

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality.
type RateLimitMiddleware struct {
	limiter   Limiter
	serviceID string
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(serviceID string, config *configv1.RateLimitConfig) (*RateLimitMiddleware, error) {
	if config == nil || !config.GetIsEnabled() {
		return nil, nil
	}

	limiter, err := createLimiter(serviceID, config)
	if err != nil {
		return nil, err
	}

	return &RateLimitMiddleware{
		limiter:   limiter,
		serviceID: serviceID,
	}, nil
}

// Execute executes the rate limiting middleware.
func (m *RateLimitMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	allowed, err := m.limiter.Allow(ctx)
	if err != nil {
		// If check fails (e.g. redis connection error).
		// Fail open.
		return next(ctx, req)
	}

	if !allowed {
		metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
			{Name: "service_id", Value: m.serviceID},
			{Name: "status", Value: "blocked"},
		})
		return nil, fmt.Errorf("rate limit exceeded for service %s", m.serviceID)
	}

	metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
		{Name: "service_id", Value: m.serviceID},
		{Name: "status", Value: "allowed"},
	})

	return next(ctx, req)
}

// Close closes the underlying limiter.
func (m *RateLimitMiddleware) Close() error {
	return m.limiter.Close()
}

func createLimiter(serviceID string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	isRedis := config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS

	if isRedis {
		return NewRedisLimiter(serviceID, config)
	}
	return &LocalLimiter{
		Limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}, nil
}
