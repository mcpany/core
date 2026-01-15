// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateClaudeConfig(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedYAML  string
		expectedError string
	}{
		{
			name: "Basic Migration",
			input: `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/path/to/files"
      ]
    }
  }
}`,
			expectedYAML: `upstreamServices:
- mcpService:
    stdioConnection:
      args:
      - -y
      - '@modelcontextprotocol/server-filesystem'
      - /path/to/files
      command: npx
  name: filesystem
`,
		},
		{
			name: "Multiple Servers with Env",
			input: `{
  "mcpServers": {
    "git": {
      "command": "docker",
      "args": ["run", "mcp/git"],
      "env": {
        "GITHUB_TOKEN": "secret"
      }
    },
    "memory": {
      "command": "uvx",
      "args": ["mcp-server-memory"]
    }
  }
}`,
			expectedYAML: `upstreamServices:
- mcpService:
    stdioConnection:
      args:
      - run
      - mcp/git
      command: docker
      env:
        GITHUB_TOKEN: secret
  name: git
- mcpService:
    stdioConnection:
      args:
      - mcp-server-memory
      command: uvx
  name: memory
`,
		},
		{
			name:          "Invalid JSON",
			input:         `{ "mcpServers": ... invalid ... }`,
			expectedError: "failed to parse Claude Desktop config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := MigrateClaudeConfig([]byte(tt.input))

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedYAML, string(output))
		})
	}
}
