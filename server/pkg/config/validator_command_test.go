// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Helper to create McpAnyServerConfig with one stdio service
func createStdioConfig(name, command string, args []string, workingDir string) *configv1.McpAnyServerConfig {
	return &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String(name),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command:          proto.String(command),
								Args:             args,
								WorkingDirectory: proto.String(workingDir),
							},
						},
					},
				},
			},
		},
	}
}

func TestValidateCommandExists_Spaces(t *testing.T) {
	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "python" || file == "node" || file == "npx" {
			return "/usr/bin/" + file, nil
		}
		return "", os.ErrNotExist
	}

	tests := []struct {
		name                 string
		command              string
		expectActionableErr  bool
		expectedSuggestion   string
	}{
		{
			name:                 "Simple command in PATH",
			command:              "python",
			expectActionableErr:  false,
		},
		{
			name:                 "Full command with args (npx)",
			command:              "npx -y mcp-server-postgres",
			expectActionableErr:  true,
			expectedSuggestion:   "It looks like you pasted a full command line into the 'command' field.",
		},
		{
			name:                 "Full command with quoted args",
			command:              `python "my script.py"`,
			expectActionableErr:  true,
			expectedSuggestion:   "Set 'command' to: \"python\"",
		},
		{
			name:                 "Command missing and no split possible",
			command:              "missing_cmd",
			expectActionableErr:  true,
			expectedSuggestion:   "Ensure \"missing_cmd\" is installed and listed in your PATH",
		},
		{
			name:                 "Command with spaces but first part also missing",
			command:              "missing -y flag",
			expectActionableErr:  true,
			expectedSuggestion:   "Ensure \"missing -y flag\" is installed", // Fallback to normal error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandExists(tt.command, "")
			if tt.expectActionableErr {
				require.Error(t, err)
				if ae, ok := err.(*ActionableError); ok {
					assert.Contains(t, ae.Suggestion, tt.expectedSuggestion)
				} else {
					t.Fatalf("Expected ActionableError, got %T: %v", err, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateStdioArgs_PackageRunners(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp(".", "mcpany-test-runners")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a script
	scriptPath := filepath.Join(tempDir, "script.py")
	err = os.WriteFile(scriptPath, []byte("print('hello')"), 0644)
	require.NoError(t, err)

	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}

	// Mock FileExists for secret validation to pass
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error { return nil }

	// Mock osStat for validating script existence
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()
	osStat = os.Stat

	tests := []struct {
		name        string
		command     string
		args        []string
		workingDir  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "npx with package name (not a file)",
			command:     "npx",
			args:        []string{"-y", "@modelcontextprotocol/server-postgres"},
			expectError: false,
		},
		{
			name:        "npx with local file",
			command:     "npx",
			args:        []string{"./script.py"},
			workingDir:  tempDir,
			expectError: false, // script.py exists in tempDir
		},
		{
			name:        "npx with missing local file",
			command:     "npx",
			args:        []string{"./missing.js"},
			workingDir:  tempDir,
			expectError: true,
			errorMsg:    "looks like a script file but does not exist",
		},
		{
			name:        "uvx with package",
			command:     "uvx",
			args:        []string{"mcp-server-postgres"},
			expectError: false,
		},
		{
			name:        "uv run with script",
			command:     "uv",
			args:        []string{"run", scriptPath},
			expectError: false,
		},
		{
			name:        "uv run with missing script",
			command:     "uv",
			args:        []string{"run", filepath.Join(tempDir, "missing.py")},
			expectError: true,
			errorMsg:    "looks like a script file but does not exist",
		},
		{
			name:        "uv run with flags then script",
			command:     "uv",
			args:        []string{"run", "--with", "requests", scriptPath},
			// Current implementation checks NEXT arg after "run"
			// If args[i+1] is flag, it might fail?
			// The code:
			// if !strings.HasPrefix(nextArg, "-") { validate... }
			// So if next arg is flag, it skips validation.
			expectError: false,
		},
		{
			name:        "bun run script",
			command:     "bun",
			args:        []string{"run", scriptPath},
			expectError: false,
		},
		{
			name:        "bun run missing script",
			command:     "bun",
			args:        []string{"run", filepath.Join(tempDir, "missing.ts")},
			expectError: true,
			errorMsg:    "looks like a script file but does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createStdioConfig("test-runner", tt.command, tt.args, tt.workingDir)
			errors := Validate(context.Background(), config, Server)

			if tt.expectError {
				require.NotEmpty(t, errors)
				assert.Contains(t, errors[0].Err.Error(), tt.errorMsg)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}
