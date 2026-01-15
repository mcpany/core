// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
)

func TestMain(m *testing.M) {
	// Mock IsSafeURL to allow all URLs during tests in this package.
	// This is necessary because many tests use httptest.NewServer which runs on localhost.
	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error { return nil }
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	os.Exit(m.Run())
}
