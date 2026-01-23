// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedCommand(t *testing.T) {
	// Helper to reset env
	resetEnv := func() {
		os.Unsetenv("MCP_ALLOWED_COMMANDS")
	}
	defer resetEnv()

	t.Run("Default (Env unset) - Allow All", func(t *testing.T) {
		resetEnv()
		assert.NoError(t, IsAllowedCommand("echo"))
		assert.NoError(t, IsAllowedCommand("rm"))
		assert.NoError(t, IsAllowedCommand("/bin/sh"))
	})

	t.Run("Env Set - Allow Specific", func(t *testing.T) {
		resetEnv()
		os.Setenv("MCP_ALLOWED_COMMANDS", "echo,ls,grep")

		assert.NoError(t, IsAllowedCommand("echo"))
		assert.NoError(t, IsAllowedCommand("ls"))
		assert.NoError(t, IsAllowedCommand("grep"))

		assert.Error(t, IsAllowedCommand("rm"))
		assert.Error(t, IsAllowedCommand("sh"))
		assert.Error(t, IsAllowedCommand("/bin/sh"))
	})

	t.Run("Env Set - Trim Whitespace", func(t *testing.T) {
		resetEnv()
		os.Setenv("MCP_ALLOWED_COMMANDS", " echo , ls ")

		assert.NoError(t, IsAllowedCommand("echo"))
		assert.NoError(t, IsAllowedCommand("ls"))
	})

	t.Run("Env Set - Empty String (Allow None?)", func(t *testing.T) {
		// If env is set but empty string, os.Getenv returns empty.
		// Our logic treats empty string from GetEnv as "unset" (Allow All).
		// If user wants to allow NONE, they must set it to something invalid or just " ".
		// os.Setenv("MCP_ALLOWED_COMMANDS", "") -> Getenv returns "" -> Allow All.
		// This is slightly ambiguous but usually Getenv behavior.
		// If key exists but value is empty.
		resetEnv()
		os.Setenv("MCP_ALLOWED_COMMANDS", "")
		// IsAllowedCommand checks: if allowedCmdsStr == "" { return nil }
		assert.NoError(t, IsAllowedCommand("echo"))

		// Allow None explicit
		os.Setenv("MCP_ALLOWED_COMMANDS", " ")
		assert.Error(t, IsAllowedCommand("echo"))
	})
}
