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
	"google.golang.org/protobuf/proto"
)

func TestPHPInjectionSecurity(t *testing.T) {
	// Setup a PHP tool configuration
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}.Build()

	// Use single quotes in the template to avoid generic shell injection blocks (like parens)
	// This forces the validator to rely on interpreter-specific checks.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := pb.Tool_builder{
		Name: proto.String("php-exec"),
	}.Build()

	// Create the tool
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Test cases with dangerous PHP functions that should be blocked
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "passthru",
			input: "passthru(\"echo vulnerable\");",
		},
		{
			name:  "shell_exec",
			input: "shell_exec(\"echo vulnerable\");",
		},
		{
			name:  "proc_open",
			input: "proc_open(\"ls\", [], $pipes);",
		},
		{
			name:  "pcntl_exec",
			input: "pcntl_exec(\"/bin/sh\", [\"-c\", \"ls\"]);",
		},
		{
			name:  "assert",
			input: "assert(\"phpinfo()\");",
		},
		{
			name:  "include",
			input: "include(\"/etc/passwd\");",
		},
		{
			name:  "include_once",
			input: "include_once(\"malicious.php\");",
		},
		{
			name:  "require_once",
			input: "require_once(\"malicious.php\");",
		},
		{
			name:  "dl",
			input: "dl(\"extension.so\");",
		},
		{
			name:  "syscall",
			input: "syscall(1);",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputMap := map[string]interface{}{
				"code": tc.input,
			}
			inputBytes, _ := json.Marshal(inputMap)

			req := &ExecutionRequest{
				ToolName: "php-exec",
				ToolInputs: inputBytes,
				DryRun:   true, // Use DryRun to trigger validation without execution
			}

			// Execute (DryRun)
			_, err := tool.Execute(context.Background(), req)

			// Assert that validation failed
			// If err is nil, it means the input was ALLOWED (vulnerability reproduced)
			// We assert Error because we WANT it to be blocked.
			assert.Error(t, err, "Expected validation error for dangerous PHP input: %s", tc.input)
			if err != nil {
				assert.Contains(t, err.Error(), "interpreter injection detected", "Error message should mention interpreter injection")
			} else {
				t.Logf("VULNERABILITY REPRODUCED: Input '%s' was allowed!", tc.input)
			}
		})
	}
}
