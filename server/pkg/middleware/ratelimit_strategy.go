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
	// Allow ...
	//
	// Summary: Executes Allow operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//
	// Returns:
	//   - bool: A boolean indicating success or status.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	//
	// ctx is the context for the request.
	// n is the n.
	//
	// Returns true if successful.
	// Returns an error if the operation fails.
	// AllowN ...
	//
	// Summary: Executes AllowN operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - n: int. An integer value.
	//
	// Returns:
	//   - bool: A boolean indicating success or status.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	//
	// rps is the rps.
	// burst is the burst.
	// Update ...
	//
	// Summary: Updates Update operation.
	//
	// Parameters:
	//   - rps: float64. The rps parameter.
	//   - burst: int. An integer value.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
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
	// Create ...
	//
	// Summary: Initializes Create operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - serviceID: serviceID. The identifier for the service.
	//   - limitScopeKey: limitScopeKey. The limitScopeKey parameter.
	//   - partitionKey: string. A string value.
	//   - config: *configv1.RateLimitConfig. The configuration object.
	//
	// Returns:
	//   - Limiter: The Limiter result.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	Create(ctx context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error)
}
