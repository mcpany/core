// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// LocalLimiter is an in-memory implementation of Limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
//
// _ is an unused parameter.
//
// Returns true if successful.
// Returns an error if the operation fails.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// _ is an unused parameter.
// n is the n.
//
// Returns true if successful.
// Returns an error if the operation fails.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// rps is the rps.
// burst is the burst.
func (l *LocalLimiter) Update(rps float64, burst int) {
	limit := rate.Limit(rps)
	if l.Limit() != limit {
		l.SetLimit(limit)
	}
	if l.Burst() != burst {
		l.SetBurst(burst)
	}
}

// LocalStrategy implements RateLimitStrategy for local in-memory rate limiting.
type LocalStrategy struct{}

// NewLocalStrategy creates a new LocalStrategy.
//
// Returns the result.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
// config holds the configuration settings.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *LocalStrategy) Create(_ context.Context, _, _, _ string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1
	}
	return &LocalLimiter{
		Limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}, nil
}

// ShouldUpdate checks if the limiter needs to be recreated.
// For local limiters, we can always just update the rate/burst in place, so we never need to recreate.
//
// limiter is the existing limiter instance.
// config is the new configuration.
//
// Returns true if the limiter should be recreated.
func (s *LocalStrategy) ShouldUpdate(_ Limiter, _ *configv1.RateLimitConfig) bool {
	return false
}
