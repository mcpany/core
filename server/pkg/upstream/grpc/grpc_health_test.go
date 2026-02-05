// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) Start() {
	m.Called()
}

func (m *MockHealthChecker) Stop() {
	m.Called()
}

func (m *MockHealthChecker) Check(ctx context.Context) health.CheckerResult {
	args := m.Called(ctx)
	return args.Get(0).(health.CheckerResult)
}

func (m *MockHealthChecker) GetRunningPeriodicCheckCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHealthChecker) IsStarted() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestCheckHealth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockChecker := new(MockHealthChecker)
		u := &Upstream{
			checker: mockChecker,
		}

		mockChecker.On("Check", mock.Anything).Return(health.CheckerResult{
			Status: health.StatusUp,
		})

		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
		mockChecker.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		mockChecker := new(MockHealthChecker)
		u := &Upstream{
			checker: mockChecker,
		}

		errMsg := "connection refused"
		mockChecker.On("Check", mock.Anything).Return(health.CheckerResult{
			Status: health.StatusDown,
			Details: map[string]health.CheckResult{
				"db": {
					Status: health.StatusDown,
					Error:  errors.New(errMsg),
				},
			},
		})

		err := u.CheckHealth(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
		// The string representation depends on how CheckerResult is formatted.
		// Based on test failure, it contains "down"
		assert.Contains(t, err.Error(), "down")
		assert.Contains(t, err.Error(), "connection refused")
		mockChecker.AssertExpectations(t)
	})

	t.Run("nil checker", func(t *testing.T) {
		u := &Upstream{
			checker: nil,
		}
		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
	})
}
