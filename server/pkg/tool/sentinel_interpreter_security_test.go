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

func TestCommandTool_InterpreterInjection_NegativeTest(t *testing.T) {
	// Negative test to ensure that we block 'system "ls"' and 'exec "cmd"' unparenthesized calls
    // when using an interpreter, which would otherwise lead to Remote Code Execution.

	toolDef := v1.Tool_builder{
		Name: proto.String("interpreter-tool"),
        InputSchema: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "properties": structpb.NewStructValue(&structpb.Struct{
                    Fields: map[string]*structpb.Value{
                        "code": structpb.NewStructValue(&structpb.Struct{}),
                    },
                }),
            },
        },
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
		Args: []string{"-e", "'{{code}}'"},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "Blocked system call without parenthesis",
			payload: `system "ls"`,
		},
		{
			name:    "Blocked open with pipe",
			payload: `open("|ls")`,
		},
		{
			name:    "Blocked backtick execution",
			payload: `exec ` + "`ls`",
		},
        {
			name:    "Blocked python exec with quotes",
			payload: `exec "import os; os.system('ls')"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputMap := map[string]interface{}{
				"code": tc.payload,
			}
			inputBytes, _ := json.Marshal(inputMap)

			req := &ExecutionRequest{
				ToolName: "interpreter-tool",
				ToolInputs: inputBytes,
			}

			_, err := localTool.Execute(context.Background(), req)

			assert.Error(t, err, "Payload %q should be blocked but was allowed", tc.payload)
			assert.Contains(t, err.Error(), "injection detected", "Payload %q should be flagged as injection", tc.payload)
		})
	}
}
