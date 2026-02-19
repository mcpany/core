// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportCmd(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		outputFile     string // if empty, stdout is used
		expectError    string // if empty, success is expected
		expectContains []string
	}{
		{
			name: "Happy Path Stdout",
			inputJSON: `{
  "mcpServers": {
    "test-server": {
      "command": "echo",
      "args": ["hello"],
      "env": {"FOO": "BAR"}
    }
  }
}`,
			expectContains: []string{
				"upstream_services:",
				"- name: test-server",
				"mcp_service:",
				"stdio_connection:",
				"command: echo",
				"args:",
				"- hello",
				"env:",
				"FOO: BAR",
			},
		},
		{
			name: "Happy Path File Output",
			inputJSON: `{
  "mcpServers": {
    "file-server": {
      "command": "cat",
      "args": ["file.txt"]
    }
  }
}`,
			outputFile: "output.yaml",
			expectContains: []string{
				"Successfully imported configuration to",
			},
		},
		{
			name:        "File Not Found",
			inputJSON:   "", // File won't be created because we pass a non-existent path
			expectError: "failed to read input file",
		},
		{
			name:        "Invalid JSON",
			inputJSON:   "invalid json content",
			expectError: "failed to parse Claude Desktop config",
		},
		{
			name: "Empty Config",
			inputJSON: `{
  "mcpServers": {}
}`,
			expectContains: []string{
				"upstream_services: []",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp input file
			var inputPath string
			if tt.name == "File Not Found" {
				inputPath = "non_existent_file.json"
			} else {
				tmpFile, err := os.CreateTemp("", "claude_config_*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(tt.inputJSON)
				require.NoError(t, err)
				err = tmpFile.Close()
				require.NoError(t, err)
				inputPath = tmpFile.Name()
			}

			// Prepare command args
			args := []string{"import", inputPath}

			// Prepare output file if needed
			var outPath string
			if tt.outputFile != "" {
				tmpDir := t.TempDir()
				outPath = filepath.Join(tmpDir, tt.outputFile)
				args = append(args, "-o", outPath)
			}

			cmd := newRootCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
				output := b.String()

				// Verify stdout content
				for _, expected := range tt.expectContains {
					if tt.outputFile != "" && strings.Contains(expected, "upstream_services") {
						// For file output, YAML content is not in stdout, but "Successfully imported" is
						continue
					}
					assert.Contains(t, output, expected)
				}

				// Verify file content if applicable
				if tt.outputFile != "" {
					content, err := os.ReadFile(outPath)
					require.NoError(t, err)
					fileContent := string(content)

					// Re-check expectations against file content
					// We need to parse the inputJSON to know what to expect in the file
					// But for simplicity, we can check basic strings derived from input
					// In a real scenario, we might want to unmarshal the result and check struct equality
					if strings.Contains(tt.inputJSON, "file-server") {
						assert.Contains(t, fileContent, "name: file-server")
						assert.Contains(t, fileContent, "command: cat")
					}
				}
			}
		})
	}
}
