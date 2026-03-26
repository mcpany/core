// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"errors"
	"testing"
)

// TestErrToolNotFound ...
// Summary: TestErrToolNotFound
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	err := ErrToolNotFound
	if err == nil {
		t.Error("Expected ErrToolNotFound to be non-nil")
	}

	if !errors.Is(err, ErrToolNotFound) {
		t.Errorf("Expected error to be ErrToolNotFound, got %v", err)
	}
}
