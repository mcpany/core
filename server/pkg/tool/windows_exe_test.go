// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsShellCommand_WindowsExe(t *testing.T) {
	// "python3" is protected
	assert.True(t, isShellCommand("python3"), "python3 should be protected")

	// "python3.10" is protected (version suffix)
	assert.True(t, isShellCommand("python3.10"), "python3.10 should be protected")

	// "python3.exe" should be protected but might fail due to .exe suffix check
	assert.True(t, isShellCommand("python3.exe"), "python3.exe should be protected")

	// "cmd.exe" is explicitly in the list, so it should pass
	assert.True(t, isShellCommand("cmd.exe"), "cmd.exe should be protected")

	// "node.exe" is NOT explicitly in the list (only "node"), so it relies on suffix check
	assert.True(t, isShellCommand("node.exe"), "node.exe should be protected")
}
