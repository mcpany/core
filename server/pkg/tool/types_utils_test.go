// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForAbsolutePath_Coverage(t *testing.T) {
	absPath, _ := filepath.Abs("test")
	err := checkForAbsolutePath(absPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute path detected")

	err = checkForAbsolutePath("relative")
	assert.NoError(t, err)
}

func TestIsSensitiveHeader_Coverage(t *testing.T) {
	assert.True(t, isSensitiveHeader("Authorization"))
	assert.True(t, isSensitiveHeader("Proxy-Authorization"))
	assert.True(t, isSensitiveHeader("Cookie"))
	assert.True(t, isSensitiveHeader("Set-Cookie"))
	assert.True(t, isSensitiveHeader("X-API-Key"))
	assert.True(t, isSensitiveHeader("my-token-header"))
	assert.True(t, isSensitiveHeader("my-secret-header"))
	assert.True(t, isSensitiveHeader("my-password-header"))
	assert.False(t, isSensitiveHeader("Content-Type"))
}
