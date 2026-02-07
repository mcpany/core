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
)

func TestLocalCommandTool_AwkLFI_Getline(t *testing.T) {
	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("awk-lfi"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: getline < "/etc/passwd"
	payload := `BEGIN { getline line < "/etc/passwd"; print line }`

	req := &ExecutionRequest{
		ToolName: "awk-lfi",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("VULNERABILITY CONFIRMED: Allowed payload %q", payload)
		t.Fail()
	} else {
		t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	}
}

func TestLocalCommandTool_PythonRCE_Import(t *testing.T) {
	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("python-rce"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: __import__('os').system('ls')
	payload := `__import__('subprocess').call(['ls'])`

	req := &ExecutionRequest{
		ToolName: "python-rce",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("VULNERABILITY CONFIRMED: Allowed payload %q", payload)
		t.Fail()
	} else {
		t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	}
}
