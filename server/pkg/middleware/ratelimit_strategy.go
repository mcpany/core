// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Limiter interface defines the methods required for a rate limiter.
//
// Summary: Represents a Limiter.
type Limiter interface {
	// Allow checks if the request is allowed.
	//
	// ctx is the context for the request.
	//
	// Returns true if successful.
	// Returns an error if the operation fails.
	//
	// Summary: Allows.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	//
	// ctx is the context for the request.
	// n is the n.
	//
	// Returns true if successful.
	// Returns an error if the operation fails.
	//
	// Summary: Allows n.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	//
	// rps is the rps.
	// burst is the burst.
	//
	// Summary: Updates.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	Update(rps float64, burst int)
}

// RateLimitStrategy defines the interface for creating rate limiters.
//
// Summary: Represents a RateLimitStrategy.
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
	//
	// Summary: Creates.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
