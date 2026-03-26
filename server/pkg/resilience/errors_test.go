// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCircuitBreakerOpenError ...
// Summary: TestCircuitBreakerOpenError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	err := &CircuitBreakerOpenError{}
	assert.Equal(t, "circuit breaker is open", err.Error())
}

// TestPermanentError ...
// Summary: TestPermanentError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Run("with error", func(t *testing.T) {
		originalErr := errors.New("original error")
		permErr := &PermanentError{Err: originalErr}
		assert.Equal(t, "original error", permErr.Error())
		assert.Equal(t, originalErr, permErr.Unwrap())
	})

	t.Run("without error", func(t *testing.T) {
		permErr := &PermanentError{Err: nil}
		assert.Equal(t, "permanent error", permErr.Error())
		assert.Nil(t, permErr.Unwrap())
	})
}
