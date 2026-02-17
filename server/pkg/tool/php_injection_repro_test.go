// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPHPInjectionRepro(t *testing.T) {
	// Configure a tool that executes PHP code
	toolProto := pb.Tool_builder{
		Name:                proto.String("run_php"),
		UnderlyingMethodFqn: proto.String("php"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call_id")

	// Test cases for dangerous functions that should be blocked but might be allowed currently
	testCases := []struct {
		name      string
		input     string
		shouldErr bool // True if we expect an error (blocked), False if allowed (vulnerable)
	}{
		{
			name:      "passthru",
			input:     "passthru(\"id\");",
			shouldErr: true, // Now blocked
		},
		{
			name:      "assert_phpinfo",
			input:     "assert(\"phpinfo()\");",
			shouldErr: true, // Now blocked
		},
		{
			name:      "include",
			input:     "include(\"/etc/passwd\");",
			shouldErr: true, // Now blocked
		},
		{
			name:      "system (already blocked)",
			input:     "system('ls');",
			shouldErr: true, // Already blocked
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := map[string]interface{}{
				"code": tc.input,
			}
			inputBytes, _ := json.Marshal(args)

			req := &ExecutionRequest{
				ToolName: "run_php",
				ToolInputs: inputBytes,
				DryRun:   true, // DryRun triggers validation without executing
			}

			_, err := tool.Execute(context.Background(), req)

			if tc.shouldErr {
				assert.Error(t, err, "Expected error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Expected no error (vulnerability repro) for input: %s", tc.input)
			}
		})
	}
}
