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

func TestLocalCommandTool_PythonInjection_Eval(t *testing.T) {
	t.Parallel()

	// Define a tool that uses 'python3' with eval()
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"val": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("python-eval-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `python3 -c "eval('{{val}}')"`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "eval('{{val}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("val")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: __import__("subprocess").run(["echo", "pwned"])
	// This uses double quotes inside single quotes.
	payload := `__import__("subprocess").run(["echo", "pwned"])`

	req := &ExecutionRequest{
		ToolName: "python-eval-tool",
		Arguments: map[string]interface{}{
			"val": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// We expect failure due to injection detection.
	// If it succeeds, it means RCE is possible.
	assert.Error(t, err, "Expected error due to Python injection detection")
	if err != nil {
		t.Logf("Error received: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	}
}
