// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

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
	// Allow loopback resources for tests
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	code := m.Run()
	os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	os.Exit(code)
}
