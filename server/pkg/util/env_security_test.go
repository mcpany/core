// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEnvVarAllowed(t *testing.T) {
	// Cleanup env after test
	defer func() {
		os.Unsetenv("MCPANY_ALLOWED_ENV")
		os.Unsetenv("MCPANY_STRICT_ENV_MODE")
	}()

	tests := []struct {
		name        string
		envVar      string
		allowedEnv  string
		strictMode  string
		wantAllowed bool
	}{
		{
			name:        "Block MCPANY_* by default",
			envVar:      "MCPANY_SECRET",
			wantAllowed: false,
		},
		{
			name:        "Allow normal var by default",
			envVar:      "MY_VAR",
			wantAllowed: true,
		},
		{
			name:        "Allow whitelisted MCPANY var",
			envVar:      "MCPANY_PUBLIC",
			allowedEnv:  "MCPANY_PUBLIC",
			wantAllowed: true,
		},
		{
			name:        "Allow wildcard whitelisted MCPANY var",
			envVar:      "MCPANY_PUBLIC_KEY",
			allowedEnv:  "MCPANY_PUBLIC_*",
			wantAllowed: true,
		},
		{
			name:        "Block normal var in Strict Mode",
			envVar:      "MY_VAR",
			strictMode:  "true",
			wantAllowed: false,
		},
		{
			name:        "Allow whitelisted var in Strict Mode",
			envVar:      "MY_VAR",
			allowedEnv:  "MY_VAR",
			strictMode:  "true",
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.allowedEnv != "" {
				os.Setenv("MCPANY_ALLOWED_ENV", tt.allowedEnv)
			} else {
				os.Unsetenv("MCPANY_ALLOWED_ENV")
			}

			if tt.strictMode != "" {
				os.Setenv("MCPANY_STRICT_ENV_MODE", tt.strictMode)
			} else {
				os.Unsetenv("MCPANY_STRICT_ENV_MODE")
			}

			allowed := IsEnvVarAllowed(tt.envVar)
			assert.Equal(t, tt.wantAllowed, allowed, "IsEnvVarAllowed(%q)", tt.envVar)
		})
	}
}
