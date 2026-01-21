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
		name            string
		input           string
		expectedFile    string
		expectedContent []string
		setup           func() string // Optional setup, returns file path if needed
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
		{
			name: "Generate SQL Config",
			// Inputs: Name, Type (3=SQL), Driver, DSN, Output Filename
			input: "my-sql-service\n3\npostgres\npostgres://user:pass@localhost:5432/db\n" + filepath.Join(tmpDir, "sql_config.yaml") + "\n",
			expectedFile: filepath.Join(tmpDir, "sql_config.yaml"),
			expectedContent: []string{
				`name: "my-sql-service"`,
				`sql_service:`,
				`driver: "postgres"`,
				`dsn: "postgres://user:pass@localhost:5432/db"`,
			},
		},
		{
			name: "Generate gRPC Config",
			// Inputs: Name, Type (4=gRPC), Address, Output Filename
			input: "my-grpc-service\n4\nlocalhost:50051\n" + filepath.Join(tmpDir, "grpc_config.yaml") + "\n",
			expectedFile: filepath.Join(tmpDir, "grpc_config.yaml"),
			expectedContent: []string{
				`name: "my-grpc-service"`,
				`grpc_service:`,
				`address: "localhost:50051"`,
			},
		},
		{
			name: "Default to HTTP on Invalid Choice",
			// Inputs: Name, Type (Invalid), Address (for HTTP fallback), Output Filename
			input: "my-fallback-service\n99\nhttps://fallback.com\n" + filepath.Join(tmpDir, "fallback_config.yaml") + "\n",
			expectedFile: filepath.Join(tmpDir, "fallback_config.yaml"),
			expectedContent: []string{
				`name: "my-fallback-service"`,
				`http_service:`,
				`address: "https://fallback.com"`,
			},
		},
		{
			name: "Overwrite Existing File - Yes",
			setup: func() string {
				path := filepath.Join(tmpDir, "overwrite_yes.yaml")
				_ = os.WriteFile(path, []byte("old content"), 0644)
				return path
			},
			// Inputs: Name, Type (1), Address, Output Filename, Overwrite (y)
			input: "my-overwrite-service\n1\nhttps://new.com\n" + filepath.Join(tmpDir, "overwrite_yes.yaml") + "\n" + "y\n",
			expectedFile: filepath.Join(tmpDir, "overwrite_yes.yaml"),
			expectedContent: []string{
				`address: "https://new.com"`,
			},
		},
		{
			name: "Overwrite Existing File - No",
			setup: func() string {
				path := filepath.Join(tmpDir, "overwrite_no.yaml")
				_ = os.WriteFile(path, []byte("old content"), 0644)
				return path
			},
			// Inputs: Name, Type (1), Address, Output Filename, Overwrite (n)
			input: "my-abort-service\n1\nhttps://new.com\n" + filepath.Join(tmpDir, "overwrite_no.yaml") + "\n" + "n\n",
			expectedFile: filepath.Join(tmpDir, "overwrite_no.yaml"),
			expectedContent: []string{
				"old content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

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
