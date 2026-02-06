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

func TestLocalCommandTool_EnvInjection_Prevention(t *testing.T) {
	t.Parallel()

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"UNTRUSTED_INPUT": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("vuln-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"), // Shell command -> enforces env injection checks
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo $UNTRUSTED_INPUT"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("UNTRUSTED_INPUT")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case 1: Malicious Input (Should be blocked)
	payload := "; echo hacked"
	req := &ExecutionRequest{
		ToolName: "vuln-tool",
		Arguments: map[string]interface{}{
			"UNTRUSTED_INPUT": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err, "Should error on malicious env input")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected in environment variable")
	}

	// Case 2: Benign Input with Spaces (Should be allowed)
	// We allow spaces in env vars to support normal sentences/paths.
	benignPayload := "Hello World"
	req2 := &ExecutionRequest{
		ToolName: "vuln-tool",
		Arguments: map[string]interface{}{
			"UNTRUSTED_INPUT": benignPayload,
		},
	}
	req2.ToolInputs, _ = json.Marshal(req2.Arguments)

	_, err2 := localTool.Execute(context.Background(), req2)
	assert.NoError(t, err2, "Should NOT error on benign env input with spaces")
}
