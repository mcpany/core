// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestPerlRCE_Unquoted(t *testing.T) {
	// Define a tool that runs perl -e {{script}}
	toolDef := v1.Tool_builder{
		Name: proto.String("perl_runner"),
	}.Build()

	cmd := "perl"
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload that avoids forbidden characters: () " ' ; etc.
	// system ls
	// in perl, system ls is system('ls') because ls is a bareword.
	payload := "system ls"

	req := &ExecutionRequest{
		ToolName: "perl_runner",
		ToolInputs: []byte(`{"script": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)
	if err != nil {
		if strings.Contains(err.Error(), "injection detected") {
			t.Logf("Security check passed: %v", err)
			return
		}
		t.Fatalf("Execution failed with unexpected error: %v", err)
	}

	t.Errorf("Security check failed: payload %q was allowed executed", payload)
}

func TestPythonRCE_Unquoted(t *testing.T) {
	// Define a tool that runs python -c {{script}}
	toolDef := v1.Tool_builder{
		Name: proto.String("python_runner"),
	}.Build()

	cmd := "python"
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload: import os
	// This should be blocked if strict mode is enabled.
	payload := "import os"

	req := &ExecutionRequest{
		ToolName: "python_runner",
		ToolInputs: []byte(`{"script": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)
	if err != nil {
		if strings.Contains(err.Error(), "injection detected") {
			t.Logf("Security check passed: %v", err)
			return
		}
		t.Fatalf("Execution failed with unexpected error: %v", err)
	}

	t.Errorf("Security check failed: payload %q was allowed executed", payload)
}
