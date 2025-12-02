// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mutex sync.Mutex

	state        State
	failures     int
	lastFailure  time.Time
	openTime     time.Time
	halfOpenHits int

	config *configv1.CircuitBreakerConfig
}

func NewCircuitBreaker(config *configv1.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

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

type CircuitBreakerOpenError struct{}

func (e *CircuitBreakerOpenError) Error() string {
	return "circuit breaker is open"
}
