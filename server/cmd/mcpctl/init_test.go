// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Override output path flag for testing
	err := cmd.Flags().Set("output", configFile)
	assert.NoError(t, err)

	// 1. Success case
	err = cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "global_settings:")
	assert.Contains(t, string(content), "example-service")

	// 2. Failure case: File already exists
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
