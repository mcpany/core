// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `
hooks:
  - path: /custom/markdown
    handler: markdown
  - path: /api/truncate
    handler: truncate
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Len(t, cfg.Hooks, 2)
	assert.Equal(t, "/custom/markdown", cfg.Hooks[0].Path)
	assert.Equal(t, "markdown", cfg.Hooks[0].Handler)
	assert.Equal(t, "/api/truncate", cfg.Hooks[1].Path)
	assert.Equal(t, "truncate", cfg.Hooks[1].Handler)
}

func TestLoadConfig_Invalid(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	content := `invalid yaml`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
}

func TestLoadConfig_Missing(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	assert.Error(t, err)
}
