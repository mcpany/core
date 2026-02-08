// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager orchestrates resilience features like circuit breakers, retries, and timeouts.
//
// Summary: orchestrates resilience features like circuit breakers, retries, and timeouts.
type Manager struct {
	circuitBreaker *CircuitBreaker
	retry          *Retry
	timeout        *Timeout
}

// NewManager creates a new Manager with the given resilience configuration.
//
// Summary: creates a new Manager with the given resilience configuration.
//
// Parameters:
//   - config: *configv1.ResilienceConfig. The config.
//
// Returns:
//   - *Manager: The *Manager.
func NewManager(config *configv1.ResilienceConfig) *Manager {
	if config == nil {
		return nil
	}

	var cb *CircuitBreaker
	if config.GetCircuitBreaker() != nil {
		cb = NewCircuitBreaker(config.GetCircuitBreaker())
	}

	var r *Retry
	if config.GetRetryPolicy() != nil {
		r = NewRetry(config.GetRetryPolicy())
	}

	var t *Timeout
	if config.GetTimeout() != nil {
		t = NewTimeout(config.GetTimeout())
	}

	if cb == nil && r == nil && t == nil {
		return nil
	}

	return &Manager{
		circuitBreaker: cb,
		retry:          r,
		timeout:        t,
	}
}

// Execute wraps the given function with resilience features.
//
// Summary: wraps the given function with resilience features.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - work: func(context.Context) error. The work.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *Manager) Execute(ctx context.Context, work func(context.Context) error) error {
	if m == nil {
		return work(ctx)
	}

	// Order of execution:
	// 1. Timeout (wraps everything else)
	// 2. Retry (retries the circuit breaker execution)
	// 3. Circuit Breaker (protects the actual call)
	//
	// Note: Timeout applies to the whole operation including retries.
	// If you want timeout per retry, the nesting would be different.
	// Typically, we want an overall timeout.

	// Apply Timeout
	if m.timeout != nil {
		return m.timeout.Execute(ctx, func(ctx context.Context) error {
			return m.executeRetryAndCB(ctx, work)
		})
	}

	return m.executeRetryAndCB(ctx, work)
}

func (m *Manager) executeRetryAndCB(ctx context.Context, work func(context.Context) error) error {
	if m.retry != nil {
		return m.retry.Execute(ctx, func(ctx context.Context) error {
			if m.circuitBreaker != nil {
				return m.circuitBreaker.Execute(ctx, work)
			}
			return work(ctx)
		})
	}

	if m.circuitBreaker != nil {
		return m.circuitBreaker.Execute(ctx, work)
	}

	return work(ctx)
}
