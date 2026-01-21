// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
	// Setup temporary directory for output files
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		input          string
		expectedFile   string
		expectedContent []string
	}{
		{
			name: "Generate HTTP Config",
			// Inputs: Name, Type (1=HTTP), Address, Output Filename
			input: "my-http-service\n1\nhttps://api.test.com\n" + filepath.Join(tmpDir, "http_config.yaml") + "\n",
			expectedFile: filepath.Join(tmpDir, "http_config.yaml"),
			expectedContent: []string{
				`name: "my-http-service"`,
				`http_service:`,
				`address: "https://api.test.com"`,
			},
		},
		{
			name: "Generate Command Config",
			// Inputs: Name, Type (2=Command), Command, Args, Output Filename
			input: "my-cli-service\n2\npython3\nmain.py --test\n" + filepath.Join(tmpDir, "cli_config.yaml") + "\n",
			expectedFile: filepath.Join(tmpDir, "cli_config.yaml"),
			expectedContent: []string{
				`name: "my-cli-service"`,
				`command_service:`,
				`command: "python3"`,
				`- "main.py"`,
				`- "--test"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin and stdout
			inputBuf := bytes.NewBufferString(tt.input)
			outputBuf := new(bytes.Buffer)

			// Execute command
			cmd := initCmd
			cmd.SetIn(inputBuf)
			cmd.SetOut(outputBuf)

			// We need to reset args to avoid interference from other tests if any
			cmd.SetArgs([]string{})

			err := cmd.Execute()
			assert.NoError(t, err)

			// Verify file creation
			assert.FileExists(t, tt.expectedFile)

			// Verify content
			content, err := os.ReadFile(tt.expectedFile)
			assert.NoError(t, err)
			contentStr := string(content)

			for _, expected := range tt.expectedContent {
				assert.Contains(t, contentStr, expected)
			}
		})
	}
}
