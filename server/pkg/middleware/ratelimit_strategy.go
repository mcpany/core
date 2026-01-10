// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Limiter interface defines the methods required for a rate limiter.
type Limiter interface {
	// Allow checks if the request is allowed.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	Update(rps float64, burst int)
}

// RateLimitStrategy defines the interface for creating rate limiters.
type RateLimitStrategy interface {
	// Create creates a new Limiter instance.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
