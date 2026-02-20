// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportCmd(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		outputFlag     string
		expectedOutput string // For stdout or console message
		expectedFile   string // Expected YAML content in output file (if outputFlag used)
		wantErr        bool
		errorMsg       string
	}{
		{
			name: "Happy Path Stdout",
			inputJSON: `{
				"mcpServers": {
					"filesystem": {
						"command": "npx",
						"args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/example/Desktop"],
						"env": {
							"NODE_ENV": "production"
						}
					}
				}
			}`,
			expectedOutput: `upstream_services:
    - name: filesystem
      mcp_service:
        stdio_connection:
            command: npx
            args:
                - -y
                - '@modelcontextprotocol/server-filesystem'
                - /Users/example/Desktop
            env:
                NODE_ENV: production

`,
			wantErr: false,
		},
		{
			name: "Happy Path File Output",
			inputJSON: `{
				"mcpServers": {
					"git": {
						"command": "docker",
						"args": ["run", "-i", "--rm", "mcp/git", "--repository", "/path/to/repo"]
					}
				}
			}`,
			outputFlag: "test_output.yaml",
			// expectedOutput will be constructed dynamically
			expectedFile: `upstream_services:
    - name: git
      mcp_service:
        stdio_connection:
            command: docker
            args:
                - run
                - -i
                - --rm
                - mcp/git
                - --repository
                - /path/to/repo
`,
			wantErr: false,
		},
		{
			name: "Input File Not Found",
			inputJSON: "", // File won't be created
			wantErr:   true,
			errorMsg:  "failed to read input file",
		},
		{
			name: "Invalid JSON",
			inputJSON: `{ "mcpServers": { ... invalid ... } }`,
			wantErr:   true,
			errorMsg:  "failed to parse Claude Desktop config",
		},
		{
			name: "Empty Config",
			inputJSON: `{ "mcpServers": {} }`,
			expectedOutput: `upstream_services: []

`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary directory for test files
			tmpDir := t.TempDir()

			var inputPath string
			if tt.name != "Input File Not Found" {
				inputFile := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(inputFile, []byte(tt.inputJSON), 0600)
				require.NoError(t, err)
				inputPath = inputFile
			} else {
				inputPath = filepath.Join(tmpDir, "nonexistent.json")
			}

			// Setup output path if needed
			var args []string
			args = append(args, inputPath)

			var outputPath string
			if tt.outputFlag != "" {
				outputPath = filepath.Join(tmpDir, tt.outputFlag)
				args = append(args, "--output", outputPath)
			}

			cmd := newImportCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)

				// Verify stdout output
				if tt.outputFlag != "" {
					// Expected output is dynamic based on temp dir path
					expected := "Successfully imported configuration to " + outputPath + "\n"
					assert.Equal(t, expected, buf.String())
				} else {
					assert.Equal(t, tt.expectedOutput, buf.String())
				}

				// Verify file output if flag was set
				if tt.outputFlag != "" {
					content, err := os.ReadFile(outputPath)
					require.NoError(t, err)
					assert.Equal(t, tt.expectedFile, string(content))
				}
			}
		})
	}
}
