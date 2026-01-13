// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateClaudeConfig(t *testing.T) {
	input := []byte(`{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/Users/user/Desktop",
        "/Users/user/Downloads"
      ]
    },
    "git": {
      "command": "uvx",
      "args": ["mcp-server-git", "--repository", "/path/to/repo"],
      "env": {
          "GITHUB_TOKEN": "my-token"
      }
    }
  }
}`)

	yamlOutput, err := MigrateClaudeConfig(input)
	require.NoError(t, err)

	// Update assertions to match the actual output format (snake_case from proto json tag)
	assert.Contains(t, yamlOutput, "upstream_services:")
	assert.Contains(t, yamlOutput, "name: filesystem")
	assert.Contains(t, yamlOutput, "command: npx")
	assert.Contains(t, yamlOutput, "- -y")
	assert.Contains(t, yamlOutput, "- '@modelcontextprotocol/server-filesystem'")
	assert.Contains(t, yamlOutput, "name: git")
	assert.Contains(t, yamlOutput, "env:")
	assert.Contains(t, yamlOutput, "plain_text: my-token")
}

func TestMigrateClaudeConfig_InvalidJSON(t *testing.T) {
	input := []byte(`{ invalid json }`)
	_, err := MigrateClaudeConfig(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Claude config")
}
