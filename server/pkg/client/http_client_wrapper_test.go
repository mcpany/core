// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// MockHealthChecker mocks the health.Checker interface.
type MockHealthChecker struct {
	Result     health.CheckerResult
	StopCalled bool
}

func (m *MockHealthChecker) Check(ctx context.Context) health.CheckerResult {
	return m.Result
}

func (m *MockHealthChecker) GetRunningPeriodicCheckCount() int {
	return 0
}

func (m *MockHealthChecker) IsStarted() bool {
	return true
}

func (m *MockHealthChecker) Start() {
}

func (m *MockHealthChecker) Stop() {
	m.StopCalled = true
}

func TestNewHTTPClientWrapper(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := &MockHealthChecker{}

	wrapper := NewHTTPClientWrapper(client, config, checker)
	assert.NotNil(t, wrapper)
	assert.Equal(t, client, wrapper.Client)
	// We can't access unexported fields config and checker directly,
	// but we can verify behavior through IsHealthy.
}

func TestHTTPClientWrapper_IsHealthy(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}

	t.Run("With Mock Checker - Up", func(t *testing.T) {
		checker := &MockHealthChecker{
			Result: health.CheckerResult{Status: health.StatusUp},
		}
		wrapper := NewHTTPClientWrapper(client, config, checker)
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("With Mock Checker - Down", func(t *testing.T) {
		checker := &MockHealthChecker{
			Result: health.CheckerResult{Status: health.StatusDown},
		}
		wrapper := NewHTTPClientWrapper(client, config, checker)
		assert.False(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("With Mock Checker - Unknown", func(t *testing.T) {
		checker := &MockHealthChecker{
			Result: health.CheckerResult{Status: health.StatusUnknown},
		}
		wrapper := NewHTTPClientWrapper(client, config, checker)
		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}

func TestHTTPClientWrapper_Close(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := &MockHealthChecker{}

	wrapper := NewHTTPClientWrapper(client, config, checker)
	err := wrapper.Close()
	assert.NoError(t, err)
	// Verify that Stop is NOT called, as HTTPClientWrapper.Close is currently a no-op regarding the checker.
	// If this behavior changes, this assertion should be updated.
	assert.False(t, checker.StopCalled)
}

func TestHTTPClientWrapper_IsHealthy_NilCheckerFallback(t *testing.T) {
	// This test verifies that if we pass nil checker, the wrapper initializes one internally.
	// We can't easily spy on the internal checker, but we can verify it doesn't panic.
	// Since the internal checker will try to make a network call to an empty config,
	// IsHealthy might return false (StatusDown) or true if default behavior is lenient?
	// Looking at code:
	// if w.checker == nil { return true } -- this is inside IsHealthy, but New sets it!
	// New: if checker == nil { checker = healthChecker.NewChecker(config) }
	// So w.checker is NOT nil.
	// healthChecker.NewChecker likely returns a checker that checks something.
	// If config is empty, what does it check?
	// This is integration behavior, but let's just ensure no panic.

	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	wrapper := NewHTTPClientWrapper(client, config, nil)
	defer func() { _ = wrapper.Close() }()

	// Should not panic
	_ = wrapper.IsHealthy(context.Background())
}
