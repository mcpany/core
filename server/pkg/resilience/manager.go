// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager orchestrates resilience features like circuit breakers and retries.
type Manager struct {
	circuitBreaker *CircuitBreaker
	retry          *Retry
}

// NewManager creates a new Manager with the given resilience configuration.
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

	if cb == nil && r == nil {
		return nil
	}

	return &Manager{
		circuitBreaker: cb,
		retry:          r,
	}
}

// Execute wraps the given function with resilience features.
func (m *Manager) Execute(work func() error) error {
	if m == nil {
		return work()
	}

	if m.retry != nil && m.circuitBreaker != nil {
		return m.retry.Execute(func() error {
			return m.circuitBreaker.Execute(work)
		})
	}

	if m.retry != nil {
		return m.retry.Execute(work)
	}

	if m.circuitBreaker != nil {
		return m.circuitBreaker.Execute(work)
	}

	return work()
}
