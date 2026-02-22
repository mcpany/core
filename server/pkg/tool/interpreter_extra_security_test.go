// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestInterpreterExtraSecurity(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		argTemplate string // The template used in Args (e.g. "-c {{code}}")
		payload     string // The user input to inject
		wantErr     string // Expected error substring
	}{
		// --- PHP Injection Tests ---
		{
			name:        "PHP system call",
			command:     "php",
			argTemplate: "-r {{code}}",
			payload:     "system('ls')",
			// Blocked as dangerous keyword "system" found (unquoted) because PHP is strict
			wantErr:     "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},
		{
			name:        "PHP exec call",
			command:     "php",
			argTemplate: "-r {{code}}",
			payload:     "exec('ls')",
			wantErr:     "interpreter injection detected: dangerous keyword \"exec\" found (unquoted)",
		},
		{
			name:        "PHP backtick execution",
			command:     "php",
			argTemplate: "-r {{code}}",
			payload:     "`ls`",
			// Blocked by generic shell injection check for unquoted context first
			wantErr:     "shell injection detected: value contains dangerous character '`'",
		},
		{
			name:        "PHP variable interpolation",
			command:     "php",
			argTemplate: "-r \"print('{{code}}');\"",
			payload:     "${system('ls')}",
			// Blocked because it contains "system" keyword
			wantErr:     "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},
		{
			name:        "PHP unquoted system",
			command:     "php",
			argTemplate: "-r {{code}}",
			payload:     "system 'ls'",
			wantErr:     "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},

		// --- Expect Injection Tests ---
		{
			name:        "Expect spawn command",
			command:     "expect",
			argTemplate: "-c {{code}}",
			payload:     "spawn /bin/sh",
			// Blocked by space check in checkUnquotedInjection because Expect is treated as Shell
			wantErr:     "shell injection detected: value contains dangerous character ' '",
		},
		{
			name:        "Expect system call",
			command:     "expect",
			argTemplate: "-c {{code}}",
			payload:     "system ls",
			wantErr:     "shell injection detected: value contains dangerous character ' '",
		},

		// --- Lua Injection Tests ---
		{
			name:        "Lua os.execute",
			command:     "lua",
			argTemplate: "-e {{code}}",
			payload:     "os.execute('ls')",
			// Blocked as dangerous keyword "os" found (unquoted)
			wantErr:     "interpreter injection detected: dangerous keyword \"os\" found (unquoted)",
		},
		{
			name:        "Lua io.popen",
			command:     "lua",
			argTemplate: "-e {{code}}",
			payload:     "io.popen('ls')",
			// Blocked as dangerous keyword "popen" found (unquoted)
			wantErr:     "interpreter injection detected: dangerous keyword \"popen\" found (unquoted)",
		},
		{
			name:        "Lua require",
			command:     "lua",
			argTemplate: "-e {{code}}",
			payload:     "require 'os'",
			wantErr:     "interpreter injection detected: dangerous keyword \"require\" found (unquoted)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := configv1.CommandLineUpstreamService_builder{
				Command: proto.String(tt.command),
			}.Build()

			// Parse argTemplate to build Args. Assume simple single arg for now.
			// e.g. "-r {{code}}" -> ["-r", "{{code}}"]
			// For simplicity in test, we construct Args manually based on template pattern
			var args []string
			if tt.command == "php" {
				// PHP usually: php -r 'code'
				// But here we test injection into the argument.
				// If template is "-r {{code}}", we assume args=["-r", "{{code}}"]
				args = []string{"-r", "{{code}}"}
				if tt.argTemplate == "-r \"print('{{code}}');\"" {
					args = []string{"-r", "print('{{code}}');"}
				}
			} else if tt.command == "expect" {
				args = []string{"-c", "{{code}}"}
			} else if tt.command == "lua" {
				args = []string{"-e", "{{code}}"}
			}

			callDef := configv1.CommandLineCallDefinition_builder{
				Args: args,
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{
							Name: proto.String("code"),
						}.Build(),
					}.Build(),
				},
			}.Build()

			toolProto := pb.Tool_builder{
				Name: proto.String("test_tool"),
			}.Build()

			toolInstance := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

			// Construct input JSON
			inputs := map[string]interface{}{
				"code": tt.payload,
			}
			inputBytes, err := json.Marshal(inputs)
			require.NoError(t, err)

			req := &ExecutionRequest{
				ToolName:   "test_tool",
				ToolInputs: inputBytes,
			}

			// Execute
			_, err = toolInstance.Execute(context.Background(), req)

			// Verify Error
			if tt.wantErr != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
