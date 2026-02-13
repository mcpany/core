// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSensitivePath(t *testing.T) {
	tests := []struct {
		path      string
		shouldErr bool
	}{
		{".env", true},
		{".env.local", true},
		{".git", true},
		{".git/config", true},
		{".ssh", true},
		{"config.yaml", false},
		{"config.json", false},
		{"mcpany.db", false},
		{"normal.txt", false},
		{"src/.env", true},
		{"src/normal.txt", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			err := IsSensitivePath(tc.path)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsAllowedPath_Sensitive(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Create a .env file
	err = os.WriteFile(".env", []byte("secret"), 0600)
	require.NoError(t, err)

	// Check IsAllowedPath
	err = IsAllowedPath(".env")
	assert.Error(t, err, "IsAllowedPath should block .env")
	assert.Contains(t, err.Error(), "restricted")

	// Create a normal file
	err = os.WriteFile("normal.txt", []byte("data"), 0600)
	require.NoError(t, err)

	err = IsAllowedPath("normal.txt")
	assert.NoError(t, err)
}
