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
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Limiter defines the interface for a rate limiter.
type Limiter interface {
	Allow(ctx context.Context) (bool, error)
	UpdateConfig(rps float64, burst int)
	Close() error
}

// LocalLimiter is a wrapper around rate.Limiter.
type LocalLimiter struct {
	limiter *rate.Limiter
}

func (l *LocalLimiter) Allow(ctx context.Context) (bool, error) {
	return l.limiter.Allow(), nil
}

func (l *LocalLimiter) UpdateConfig(rps float64, burst int) {
	l.limiter.SetLimit(rate.Limit(rps))
	l.limiter.SetBurst(burst)
}

func (l *LocalLimiter) Close() error {
	return nil
}

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality for upstream services.
type RateLimitMiddleware struct {
	toolManager        tool.ManagerInterface
	limiters           sync.Map // map[string]Limiter
	redisClientFactory func(*redis.Options) *redis.Client
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		toolManager:        toolManager,
		redisClientFactory: redis.NewClient,
	}
}

// SetRedisClientFactoryForTest sets the redis client factory for testing purposes.
func (m *RateLimitMiddleware) SetRedisClientFactoryForTest(factory func(*redis.Options) *redis.Client) {
	m.redisClientFactory = factory
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
		return nil, fmt.Errorf("failed to get rate limiter: %w", err)
	}

	allowed, err := limiter.Allow(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limiting error: %w", err)
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

	// Check if existing limiter is valid and correct type
	if ok {
		existingLimiter := limiterVal.(Limiter)
		isRedisConfig := config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS
		_, isRedisLimiter := existingLimiter.(*RedisLimiter)

		if isRedisConfig == isRedisLimiter {
			// Same type, update and return
			existingLimiter.UpdateConfig(rps, burst)
			return existingLimiter, nil
		}

		// Type mismatch, proceed to create new and swap
	}

	// Create new limiter
	var newLimiter Limiter
	if config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS {
		redisConfig := config.GetRedis()
		if redisConfig == nil {
			return nil, fmt.Errorf("redis config missing for redis rate limiter")
		}
		client := m.redisClientFactory(&redis.Options{
			Addr:     redisConfig.GetAddress(),
			Password: redisConfig.GetPassword(),
			DB:       int(redisConfig.GetDb()),
		})
		newLimiter = NewRedisLimiter(client, serviceID, rps, burst)
	} else {
		newLimiter = &LocalLimiter{
			limiter: rate.NewLimiter(rate.Limit(rps), burst),
		}
	}

	previous, loaded := m.limiters.Swap(serviceID, newLimiter)
	if loaded {
		_ = previous.(Limiter).Close()
	}
	return newLimiter, nil
}
