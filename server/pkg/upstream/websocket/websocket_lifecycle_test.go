// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"context"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockChecker implements health.Checker for testing.
type MockChecker struct {
	status     health.AvailabilityStatus
	stopCalled bool
}

func (m *MockChecker) Check(_ context.Context) health.CheckerResult {
	return health.CheckerResult{
		Status: m.status,
	}
}

func (m *MockChecker) Start() {}
func (m *MockChecker) Stop()  { m.stopCalled = true }

func (m *MockChecker) GetRunningPeriodicCheckCount() int {
	return 0
}

func (m *MockChecker) IsStarted() bool {
	return true
}

func TestUpstream_CheckHealth(t *testing.T) {
	t.Run("checker is nil", func(t *testing.T) {
		u := &Upstream{}
		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
	})

	t.Run("checker returns up", func(t *testing.T) {
		u := &Upstream{
			checker: &MockChecker{status: health.StatusUp},
		}
		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
	})

	t.Run("checker returns down", func(t *testing.T) {
		u := &Upstream{
			checker: &MockChecker{status: health.StatusDown},
		}
		err := u.CheckHealth(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})
}

func TestUpstream_Shutdown(t *testing.T) {
	t.Run("shutdown success", func(t *testing.T) {
		pm := pool.NewManager()
		// Register a dummy pool to verify deregistration
		serviceID := "test-service"
		pm.Register(serviceID, "dummy-pool")

		mockChecker := &MockChecker{status: health.StatusUp}
		u := &Upstream{
			poolManager: pm,
			serviceID:   serviceID,
			checker:     mockChecker,
		}

		err := u.Shutdown(context.Background())
		require.NoError(t, err)

		// Verify pool is deregistered
		// We can't directly check internal map, but we can try to Get and fail?
		// Or assume if no panic/error it's fine.
		// Since pool.Manager.Deregister is void, we mainly check if Shutdown runs without error
		// and calls Stop on checker.
		assert.True(t, mockChecker.stopCalled, "Checker.Stop() should be called")
	})

	t.Run("checker is nil", func(t *testing.T) {
		pm := pool.NewManager()
		u := &Upstream{
			poolManager: pm,
			serviceID:   "test-service",
			checker:     nil,
		}

		err := u.Shutdown(context.Background())
		require.NoError(t, err)
	})
}
