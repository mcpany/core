// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: Limiter interface defines the methods required for a rate limiter.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package middleware

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

type Limiter interface {
	// Allow checks if the request is allowed.
	//
	// ctx is the context for the request.
	//
	// Returns true if successful.
	// Returns an error if the operation fails.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	//
	// ctx is the context for the request.
	// n is the n.
	//
	// Returns true if successful.
	// Returns an error if the operation fails.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	//
	// rps is the rps.
	// burst is the burst.
	// Summary: RateLimitStrategy defines the interface for creating rate limiters.
	//
	//
	// Errors:
	//   - An error if it fails.
	//
	// Side Effects:
	//   - None.
	Update(rps float64, burst int)
}

type RateLimitStrategy interface {
	// Create creates a new Limiter instance.
	//
	// ctx is the context for the request.
	// serviceID is the serviceID.
	// limitScopeKey is the limitScopeKey.
	// partitionKey is the partitionKey.
	// config holds the configuration settings.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
