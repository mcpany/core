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

// MockHealthChecker ...
// Summary: MockHealthChecker
	mock.Mock
}

// Start ...
// Summary: Start
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called()
}

// Stop ...
// Summary: Stop
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called()
}

// Check ...
// Summary: Check
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	return args.Get(0).(health.CheckerResult)
}

// GetRunningPeriodicCheckCount ...
// Summary: GetRunningPeriodicCheckCount
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Int(0)
}

// IsStarted ...
// Summary: IsStarted
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Bool(0)
}

// TestCheckHealth ...
// Summary: TestCheckHealth
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
