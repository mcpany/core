// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
// NewManager creates a new Manager with the given resilience configuration.
//
// Summary: Initializes a new Resilience Manager.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Parameters:
//   - config: *configv1.ResilienceConfig. The resilience configuration.
//
// Returns:
//   - *Manager: The initialized manager, or nil if no resilience features are enabled.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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

// Execute wraps the given function with resilience features.
//
// Summary: Executes the work function with configured resilience policies (timeout, retry, circuit breaker).
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - work: func(context.Context) error. The operation to execute.
//
// Returns:
//   - error: An error if the operation fails after all resilience attempts.
//
// Side Effects:
//   - Applies timeout context.
//   - Retries operation on failure.
//   - Checks and updates circuit breaker state.
// Errors:
//   - triggers relevant error states on failure.
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
