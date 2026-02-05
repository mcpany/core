// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"os"

	"github.com/mcpany/core/server/pkg/validation"
)

func init() {
	// Allow local IPs for all tests in tool_test package
	// This is necessary because many tests use httptest.NewServer which runs on 127.0.0.1.
	// This mirrors the setup in main_test.go for package tool.
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	validation.IsSafeURL = func(urlStr string) error { return nil }
}
