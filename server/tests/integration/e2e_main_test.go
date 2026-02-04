// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Enable local IP access for all integration tests to allow
	// in-process server to connect to local mock services.
	// We set this globally to avoid race conditions with t.Setenv in parallel tests.
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	os.Exit(m.Run())
}
