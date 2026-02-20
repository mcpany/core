// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// LocalLimiter is an in-memory implementation of Limiter based on token bucket.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
//
// Parameters:
//  _ (context.Context): Unused context.
//
// Returns:
//  bool: True if allowed, false otherwise.
//  error: Always nil.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Parameters:
//  _ (context.Context): Unused context.
//  n (int): The cost of the request.
//
// Returns:
//  bool: True if allowed, false otherwise.
//  error: Always nil.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update dynamically updates the limiter configuration.
//
// Parameters:
//  rps (float64): The new requests per second limit.
//  burst (int): The new burst capacity.
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
// Returns:
//  *LocalStrategy: The initialized strategy.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter instance.
//
// Parameters:
//  _ (context.Context): Unused context.
//  _ (string): Unused service ID.
//  _ (string): Unused scope key.
//  _ (string): Unused partition key.
//  config (*configv1.RateLimitConfig): The rate limit configuration.
//
// Returns:
//  Limiter: The created local limiter.
//  error: Always nil.
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
