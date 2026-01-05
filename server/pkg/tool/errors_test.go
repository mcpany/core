// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"errors"
	"testing"
)

func TestErrToolNotFound(t *testing.T) {
	t.Parallel()
	err := ErrToolNotFound
	if err == nil {
		t.Error("Expected ErrToolNotFound to be non-nil")
	}

	if !errors.Is(err, ErrToolNotFound) {
		t.Errorf("Expected error to be ErrToolNotFound, got %v", err)
	}
}
