// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import "os"

func init() {
	// Allow local IPs for testing (bypass SSRF protection) for all tests in this package
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
}
