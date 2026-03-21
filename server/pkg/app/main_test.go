// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Enable file configuration for all tests in this package
	// as many legacy tests rely on loading config from files.
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}
