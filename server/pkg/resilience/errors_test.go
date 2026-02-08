package resilience

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerOpenError(t *testing.T) {
	err := &CircuitBreakerOpenError{}
	assert.Equal(t, "circuit breaker is open", err.Error())
}

func TestPermanentError(t *testing.T) {
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
