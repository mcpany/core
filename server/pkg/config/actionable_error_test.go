// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapActionableError(t *testing.T) {
	t.Run("Wraps normal error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := WrapActionableError("context", baseErr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context: base error")

		_, ok := err.(*ActionableError)
		assert.False(t, ok, "Should not return ActionableError for normal error")
	})

	t.Run("Wraps ActionableError and preserves type", func(t *testing.T) {
		baseErr := &ActionableError{
			Err:        errors.New("base error"),
			Suggestion: "Do something",
		}
		err := WrapActionableError("context", baseErr)

		assert.Error(t, err)

		ae, ok := err.(*ActionableError)
		assert.True(t, ok, "Should return ActionableError")
		assert.Equal(t, "Do something", ae.Suggestion)

		// Check that context is preserved
		assert.Contains(t, ae.Err.Error(), "context")
		assert.Contains(t, ae.Err.Error(), "base error")
	})

	t.Run("Wraps wrapped ActionableError and preserves type and context", func(t *testing.T) {
		// Create an error chain: wrapped -> ActionableError
		innerAE := &ActionableError{
			Err:        errors.New("inner error"),
			Suggestion: "Fix it",
		}
		// Simulate intermediate wrapping (e.g. by some other function using fmt.Errorf)
		intermediateErr := fmt.Errorf("intermediate: %w", innerAE)

		// Now wrap with WrapActionableError
		err := WrapActionableError("outer", intermediateErr)

		assert.Error(t, err)

		ae, ok := err.(*ActionableError)
		assert.True(t, ok, "Should return ActionableError")
		assert.Equal(t, "Fix it", ae.Suggestion)

		// Verify FULL context is preserved
		errStr := ae.Err.Error()
		assert.Contains(t, errStr, "outer")
		assert.Contains(t, errStr, "intermediate")
		assert.Contains(t, errStr, "inner error")
	})
}
