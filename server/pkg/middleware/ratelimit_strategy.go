// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Limiter interface defines the methods required for a rate limiter.
//
// Summary: interface defines the methods required for a rate limiter.
type Limiter interface {
	// Allow checks if the request is allowed.
	//
	// Summary: checks if the request is allowed.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//
	// Returns:
	//   - bool: The bool.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	//
	// Summary: checks if the request is allowed with a specific cost.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - n: int. The int.
	//
	// Returns:
	//   - bool: The bool.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	//
	// Summary: updates the limiter configuration.
	//
	// Parameters:
	//   - rps: float64. The float64.
	//   - burst: int. The int.
	//
	// Returns:
	//   None.
	Update(rps float64, burst int)
}

// RateLimitStrategy defines the interface for creating rate limiters.
//
// Summary: defines the interface for creating rate limiters.
type RateLimitStrategy interface {
	// Create creates a new Limiter instance.
	//
	// Summary: creates a new Limiter instance.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - serviceID: string. The string.
	//   - limitScopeKey: string. The string.
	//   - partitionKey: string. The string.
	//   - config: *configv1.RateLimitConfig. The configuration.
	//
	// Returns:
	//   - Limiter: The Limiter.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
