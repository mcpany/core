// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("closed_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		cb := NewCircuitBreaker(config)
		err := cb.Execute(func() error { return nil })
		require.NoError(t, err)
		require.Equal(t, StateClosed, cb.state)
	})

	t.Run("open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Second))
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		cb.Execute(func() error { return errors.New("error") })
		cb.Execute(func() error { return errors.New("error") })

		require.Equal(t, StateOpen, cb.state)

		// Third request should be blocked
		err := cb.Execute(func() error { return nil })
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
	})

	t.Run("half_open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		cb.Execute(func() error { return errors.New("error") })
		cb.Execute(func() error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// First request in half-open state should be allowed
		err := cb.Execute(func() error { return nil })
		require.NoError(t, err)
		require.Equal(t, StateClosed, cb.state)
	})

	t.Run("half_open_to_open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		halfOpenRequests := int32(1)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		config.SetHalfOpenRequests(halfOpenRequests)
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		cb.Execute(func() error { return errors.New("error") })
		cb.Execute(func() error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// First request in half-open state should be allowed, but fail
		err := cb.Execute(func() error { return errors.New("error") })
		require.Error(t, err)
		require.Equal(t, StateOpen, cb.state)
	})
}
