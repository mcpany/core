// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_SedSandboxBypass_Env(t *testing.T) {
	// Check if sed and env are available
	if _, err := exec.LookPath("sed"); err != nil {
		t.Skip("sed not found")
	}
	if _, err := exec.LookPath("env"); err != nil {
		t.Skip("env not found")
	}

	// Create a temporary file in CWD
	fileName := "test_bypass.txt"
	err := os.WriteFile(fileName, []byte("hello"), 0644)
	require.NoError(t, err)
	defer os.Remove(fileName)

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"file": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("sed-bypass"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("env"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sed", "-e", "s/hello/pwned/", "{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "sed-bypass",
		Arguments: map[string]interface{}{
			"file": fileName, // Relative path
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail now
	result, err := localTool.Execute(context.Background(), req)

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sed tool detected behind wrapper \"env\"; execution blocked for security")
	assert.Nil(t, result)
}

func TestLocalCommandTool_SedSandboxBypass_Env_WithVars(t *testing.T) {
	// Check if sed and env are available
	if _, err := exec.LookPath("sed"); err != nil {
		t.Skip("sed not found")
	}
	if _, err := exec.LookPath("env"); err != nil {
		t.Skip("env not found")
	}

	// Create a temporary file in CWD
	fileName := "test_bypass_vars.txt"
	err := os.WriteFile(fileName, []byte("hello"), 0644)
	require.NoError(t, err)
	defer os.Remove(fileName)

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"file": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("sed-bypass-vars"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("env"),
		Local:   proto.Bool(true),
	}.Build()

	// Try to use environment variable assignment to confuse the parser
	// env FOO=BAR sed -e ...
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"FOO=BAR", "sed", "-e", "s/hello/pwned/", "{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "sed-bypass-vars",
		Arguments: map[string]interface{}{
			"file": fileName,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail now
	result, err := localTool.Execute(context.Background(), req)

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sed tool detected behind wrapper \"env\"; execution blocked for security")
	assert.Nil(t, result)
}

func TestLocalCommandTool_Sed_SandboxEnforcement(t *testing.T) {
	// Check if sed is available
	if _, err := exec.LookPath("sed"); err != nil {
		t.Skip("sed not found")
	}

	// Create a temporary file in CWD
	fileName := "test_sandbox.txt"
	err := os.WriteFile(fileName, []byte("hello"), 0644)
	require.NoError(t, err)
	defer os.Remove(fileName)

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"file": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("sed-sandbox"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sed"),
		Local:   proto.Bool(true),
	}.Build()

	// Try to use 'e' command which is forbidden in sandbox
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "s/^/pwned/e", "{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "sed-sandbox",
		Arguments: map[string]interface{}{
			"file": fileName,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail because sed returns error (illegal command)
	result, err := localTool.Execute(context.Background(), req)

	if err == nil {
		// If it succeeded, check output. If 'e' was ignored, it might succeed?
		// But if 'e' executed, we get pwned.
		resMap, _ := result.(map[string]interface{})
		stdout, _ := resMap["stdout"].(string)
		assert.NotContains(t, stdout, "pwned", "Sandbox failed to prevent execution")
	} else {
		// Expected error from sed
		// "invalid command code e" or similar
		// But we should assert it's NOT the wrapper block error
		assert.NotContains(t, err.Error(), "execution blocked for security")
	}
}
