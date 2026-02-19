// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_InterpreterKeywordCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		command     string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Perl: Safe String Literal",
			command:     "perl",
			input:       "print \"system\"",
			shouldError: false,
		},
		{
			name:        "Perl: Variable Bypass Attempt",
			command:     "perl",
			input:       "$c=\"echo pwned\"; system $c",
			shouldError: true,
			errorMsg:    "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},
		{
			name:        "Awk: System Block",
			command:     "awk",
			input:       "BEGIN { system(\"ls\") }",
			shouldError: true,
			errorMsg:    "awk injection detected: value contains 'system'",
		},
		{
			name:        "Awk: String Literal Block",
			command:     "awk",
			input:       "BEGIN { print \"system\" }",
			shouldError: true,
			errorMsg:    "awk injection detected: value contains 'system'",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			inputSchema := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"properties": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"code": structpb.NewStructValue(&structpb.Struct{}),
						},
					}),
				},
			}
			tool := v1.Tool_builder{
				Name:        proto.String("test-tool"),
				InputSchema: inputSchema,
			}.Build()

			service := configv1.CommandLineUpstreamService_builder{
				Command: proto.String(tc.command),
				Local:   proto.Bool(true),
			}.Build()

			// Configured to run `cmd '{{code}}'`
			callDef := configv1.CommandLineCallDefinition_builder{
				Args: []string{"'{{code}}'"}, // We wrap in single quotes to trigger quoteLevel=2 logic
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
					}.Build(),
				},
			}.Build()

			// Awk might need args like -e or just script?
			// The test logic in types.go checks for interpreter name.
			// So command name matters.
			// But for awk, we might need a dummy arg structure to make it runnable if we were actually running it.
			// Since LocalCommandTool.Execute does checks BEFORE execution, this is fine for unit testing the security checks.

			localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

			req := &ExecutionRequest{
				ToolName: "test-tool",
				Arguments: map[string]interface{}{
					"code": []interface{}{tc.input},
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			_, err := localTool.Execute(context.Background(), req)

			if tc.shouldError {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
