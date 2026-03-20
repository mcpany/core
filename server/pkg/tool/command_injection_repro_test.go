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

func TestLocalCommandTool_ShellInjection_ArgsBypass(t *testing.T) {
	t.Parallel()
	// Define a tool that uses 'sh', which is in isShellCommand list.
	// We allow 'args' input.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("sh-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `sh -c`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c"},
		Parameters: []*configv1.CommandLineParameterMapping{
			// Explicitly allowing args
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// We attempt to pass a command with dangerous characters (space, semicolon) via 'args'
	// 'echo hello; echo pwned'
	// checkForShellInjection blocks spaces and semicolons.
	req := &ExecutionRequest{
		ToolName: "sh-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"echo hello; echo pwned"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// Expect failure due to shell injection detection
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Nil(t, result)
}

func TestMyPythonQuotedExecfileInjection(t *testing.T) {
	val := "execfile('hack.py')"
	err := checkForShellInjection(val, "\"{{val}}\"", "{{val}}", "python", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: execfile passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}

func TestJSFunctionKeywordsInjection(t *testing.T) {
	val := "setTimeout('console.log(1)')"
	err := checkForShellInjection(val, "\"{{val}}\"", "{{val}}", "node", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: setTimeout passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}

func TestPHPKeywordsInjection(t *testing.T) {
    val := "passthru('ls')"
	err := checkForShellInjection(val, "\"{{val}}\"", "{{val}}", "php", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: passthru passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}
