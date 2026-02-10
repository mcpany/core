// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Limiter interface defines the methods required for a rate limiter.
//
// Summary: Interface for rate limiting logic.
type Limiter interface {
	// Allow checks if the request is allowed.
	//
	// Summary: Checks if a request with cost 1 is allowed.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//
	// Returns:
	//   - bool: True if allowed, false otherwise.
	//   - error: An error if the check fails.
	Allow(ctx context.Context) (bool, error)

	// AllowN checks if the request is allowed with a specific cost.
	//
	// Summary: Checks if a request with cost N is allowed.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - n: int. The cost of the request.
	//
	// Returns:
	//   - bool: True if allowed, false otherwise.
	//   - error: An error if the check fails.
	AllowN(ctx context.Context, n int) (bool, error)

	// Update updates the limiter configuration.
	//
	// Summary: Dynamically updates the limiter settings.
	//
	// Parameters:
	//   - rps: float64. Requests per second.
	//   - burst: int. Burst size.
	Update(rps float64, burst int)
}

// RateLimitStrategy defines the interface for creating rate limiters.
//
// Summary: Factory interface for rate limiters.
type RateLimitStrategy interface {
	// Create creates a new Limiter instance.
	//
	// Summary: Creates a new Limiter.
	//
	// Parameters:
	//   - ctx: context.Context. The context.
	//   - serviceID: string. The service ID.
	//   - limitScopeKey: string. The scope key (e.g. "tool:myTool").
	//   - partitionKey: string. The partition key (e.g. "ip:1.2.3.4").
	//   - config: *configv1.RateLimitConfig. The rate limit configuration.
	//
	// Returns:
	//   - Limiter: The created limiter.
	//   - error: An error if creation fails.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
