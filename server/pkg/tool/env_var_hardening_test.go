// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestDangerousEnvVars(t *testing.T) {
	// Vulnerability: Missing dangerous environment variables in blocklist.
	// Specifically GCONV_PATH (glibc RCE) and SHELL/HOME (config/execution hijacking).

	vars := []string{
		"GCONV_PATH",
		"SHELL",
		"HOME",
		"XDG_CONFIG_HOME",
	}

	for _, v := range vars {
		if isDangerousEnvVar(v) {
			t.Logf("Environment variable %q is already blocked (Good)", v)
		} else {
			t.Fatalf("Vulnerability confirmed: Environment variable %q is NOT blocked", v)
		}
	}
}
