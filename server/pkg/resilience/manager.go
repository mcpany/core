// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Manager struct {
	circuitBreaker *CircuitBreaker
	retry          *Retry
	timeout        *Timeout
}

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
