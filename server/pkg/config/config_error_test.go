// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReproduction_SilentFailure_BadConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a malformed YAML file
	_ = afero.WriteFile(fs, "bad.yaml", []byte("upstream_services:\n  - name: test\n    http_service:\n      address: http://localhost\n  indentation_error"), 0644)

	// Use the lenient store, as used in server.go
	store := NewFileStoreWithSkipErrors(fs, []string{"bad.yaml"})

	// Load should NOT return an error, effectively swallowing the parse error
	cfg, err := store.Load(context.Background())

	// We expect NO error because skipErrors=true
	assert.NoError(t, err, "Load should not return error when skipErrors is true")

	// And config should be empty or partial
	assert.Empty(t, cfg.GetUpstreamServices(), "Config should be empty because the file failed to parse")
}

func TestReproduction_ClaudeConfig_HelpfulError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a file mimicking Claude Desktop config
	claudeConfig := `
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/me/Desktop"]
    }
  }
}
`
	_ = afero.WriteFile(fs, "claude_config.json", []byte(claudeConfig), 0644)

	// 1. Test with Lenient Store (Current Behavior)
	storeLenient := NewFileStoreWithSkipErrors(fs, []string{"claude_config.json"})
	_, err := storeLenient.Load(context.Background())
	assert.NoError(t, err, "Lenient store should swallow the error")

	// 2. Test with Strict Store (Desired Behavior)
	storeStrict := NewFileStore(fs, []string{"claude_config.json"})
	_, errStrict := storeStrict.Load(context.Background())

	assert.Error(t, errStrict)
	// Verify the helpful message is present
	assert.True(t, strings.Contains(errStrict.Error(), "Did you mean \"upstream_services\"?"), "Error message should contain helpful hint")
}
