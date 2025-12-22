// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// State represents the current state of the circuit breaker.
type State int

const (
	// StateClosed represents the state where the circuit breaker allows requests to pass through.
	StateClosed State = iota
	// StateOpen represents the state where the circuit breaker blocks requests immediately.
	StateOpen
	// StateHalfOpen represents the state where the circuit breaker allows a limited number of requests to test if the service has recovered.
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern. It prevents the
// application from performing operations that are likely to fail.
type CircuitBreaker struct {
	mutex sync.Mutex

	state        State
	failures     int
	openTime     time.Time
	halfOpenHits int

	config *configv1.CircuitBreakerConfig
}

// NewCircuitBreaker creates a new CircuitBreaker with the given configuration.
// Returns the result.
func NewCircuitBreaker(config *configv1.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute runs the provided work function. If the circuit breaker is open, it
// returns a CircuitBreakerOpenError immediately. If the work function fails,
// it tracks the failure and may trip the breaker.
func (cb *CircuitBreaker) Execute(work func() error) error {
	cb.mutex.Lock()

	if cb.state == StateOpen {
		if time.Since(cb.openTime) > cb.config.GetOpenDuration().AsDuration() {
			cb.state = StateHalfOpen
			cb.halfOpenHits = 0
		} else {
			cb.mutex.Unlock()
			return &CircuitBreakerOpenError{}
		}
	}

	if cb.state == StateHalfOpen {
		if cb.halfOpenHits >= int(cb.config.GetHalfOpenRequests()) {
			cb.mutex.Unlock()
			return &CircuitBreakerOpenError{}
		}
		cb.halfOpenHits++
	}

	cb.mutex.Unlock()

	err := work()
	if err != nil {
		var permanentErr *PermanentError
		if errors.As(err, &permanentErr) {
			return err
		}

		cb.mutex.Lock()
		cb.onFailure()
		cb.mutex.Unlock()
		return err
	}

	cb.mutex.Lock()
	cb.onSuccess()
	cb.mutex.Unlock()
	return nil
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.halfOpenHits = 0
	}
	cb.failures = 0
}

func (cb *CircuitBreaker) onFailure() {
	if cb.state == StateOpen {
		return
	}

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.openTime = time.Now()
		return
	}

	cb.failures++
	if cb.failures >= int(cb.config.GetConsecutiveFailures()) {
		cb.state = StateOpen
		cb.openTime = time.Now()
	}
}

// CircuitBreakerOpenError is returned when the circuit breaker is in the Open state.
type CircuitBreakerOpenError struct{}

// Error returns the error message for a CircuitBreakerOpenError.
func (e *CircuitBreakerOpenError) Error() string {
	return "circuit breaker is open"
}
