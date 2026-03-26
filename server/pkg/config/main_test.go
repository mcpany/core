// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"
)

// TestMain ...
// Summary: TestMain
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create dummy cert file for tests
	_ = os.WriteFile("dummy-cert.pem", []byte("cert"), 0600)
	code := m.Run()
	_ = os.Remove("dummy-cert.pem")
	os.Exit(code)
}
