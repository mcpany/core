// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// State represents the current state of the circuit breaker.
type State int32

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

	state        State // Accessed using atomics for read optimization
	failures     int
	openTime     time.Time
	halfOpenHits int

	config *configv1.CircuitBreakerConfig
}

// NewCircuitBreaker creates a new CircuitBreaker with the given configuration.
func NewCircuitBreaker(config *configv1.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute runs the provided work function. If the circuit breaker is open, it
// returns a CircuitBreakerOpenError immediately. If the work function fails,
// it tracks the failure and may trip the breaker.
func (cb *CircuitBreaker) Execute(ctx context.Context, work func(context.Context) error) error {
	// Optimization: Optimistically check if Closed without lock.
	// This covers the "Happy Path" (most common case).
	if cb.getState() != StateClosed {
		// Slow path: acquire lock to check Open/HalfOpen state
		cb.mutex.Lock()

		// Re-check state under lock
		currentState := cb.getState()

		if currentState == StateOpen {
			if time.Since(cb.openTime) > cb.config.GetOpenDuration().AsDuration() {
				cb.setState(StateHalfOpen)
				cb.halfOpenHits = 0
				currentState = StateHalfOpen
			} else {
				cb.mutex.Unlock()
				return &CircuitBreakerOpenError{}
			}
		}

		if currentState == StateHalfOpen {
			if cb.halfOpenHits >= int(cb.config.GetHalfOpenRequests()) {
				cb.mutex.Unlock()
				return &CircuitBreakerOpenError{}
			}
			cb.halfOpenHits++
		}

		cb.mutex.Unlock()
	}

	err := work(ctx)
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

// getState reads the state atomically.
func (cb *CircuitBreaker) getState() State {
	return State(atomic.LoadInt32((*int32)(&cb.state)))
}

// setState updates the state atomically. Caller must hold the mutex.
func (cb *CircuitBreaker) setState(newState State) {
	atomic.StoreInt32((*int32)(&cb.state), int32(newState))
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.getState() == StateHalfOpen {
		cb.setState(StateClosed)
		cb.halfOpenHits = 0
	}
	cb.failures = 0
}

func (cb *CircuitBreaker) onFailure() {
	currentState := cb.getState()
	if currentState == StateOpen {
		return
	}

	if currentState == StateHalfOpen {
		cb.setState(StateOpen)
		cb.openTime = time.Now()
		return
	}

	cb.failures++
	if cb.failures >= int(cb.config.GetConsecutiveFailures()) {
		cb.setState(StateOpen)
		cb.openTime = time.Now()
	}
}

// CircuitBreakerOpenError is returned when the circuit breaker is in the Open state.
type CircuitBreakerOpenError struct{}

// Error returns the error message for a CircuitBreakerOpenError.
func (e *CircuitBreakerOpenError) Error() string {
	return "circuit breaker is open"
}
