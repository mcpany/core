// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Execute(t *testing.T) {
	t.Parallel()
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
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"hello", "world"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello world\n", resultMap["stdout"])

	assert.NotNil(t, localTool.Tool())
	assert.Equal(t, tool, localTool.Tool())
	assert.Nil(t, localTool.GetCacheConfig())
}

func TestLocalCommandTool_Execute_WithEnv(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-tool-env"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		Env: map[string]*configv1.SecretValue{
			"MY_ENV": configv1.SecretValue_builder{
				PlainText: proto.String("secret_value"),
			}.Build(),
		},
	}.Build()
	// Verify environment variable is passed correctly without leaking it in the output (if we were to echo it)
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "if [ \"$MY_ENV\" = \"secret_value\" ]; then echo -n match; else echo -n mismatch; fi"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-env",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "match", resultMap["stdout"])
}

func TestLocalCommandTool_Execute_RedactsSecrets(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-tool-redact"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		Env: map[string]*configv1.SecretValue{
			"MY_SECRET": configv1.SecretValue_builder{
				PlainText: proto.String("SuperSecret123"),
			}.Build(),
		},
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo -n $MY_SECRET"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-redact",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "[REDACTED]", resultMap["stdout"])
}

func TestLocalCommandTool_Execute_BlockedByPolicy(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-blocked"),
		Description: proto.String("A test tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	policies := []*configv1.CallPolicy{
		configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_DENY.Enum(),
		}.Build(),
	}

	localTool := NewLocalCommandTool(tool, service, callDef, policies, "blocked-call-id")

	req := &ExecutionRequest{
		ToolName: "test-tool-blocked",
		Arguments: map[string]interface{}{
			"args": []interface{}{"should", "not", "run"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution blocked by policy")
	assert.Nil(t, result)
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestLocalCommandTool_Execute_JSONProtocol_StderrCapture(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-json-stderr"),
		Description: proto.String("A test tool that fails"),
	}.Build()
	// Command that writes to stderr and exits with error, producing invalid JSON (empty stdout)
	service := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'something went wrong' >&2; exit 1"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-json-stderr",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)

	// This assertion should fail before fix
	assert.Contains(t, err.Error(), "something went wrong")
}

func TestLocalCommandTool_Execute_BypassLocalFileCheck(t *testing.T) {
	// Reproduction of vulnerability where setting container_environment on a local tool
	// bypasses checkForLocalFileAccess but still executes locally.

	// Setup
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
		Name:        proto.String("test-bypass"),
		InputSchema: inputSchema,
	}.Build()

	// Configuration: local=true BUT container_environment is set
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"), // Dummy command
		Local:   proto.Bool(true),
		ContainerEnvironment: configv1.ContainerEnvironment_builder{
			Image: proto.String("alpine:latest"), // Triggers isDocker = true (previously)
		}.Build(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build(),
			}.Build(),
		},
		Args: []string{"{{file}}"},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "bypass-call")

	// Create an absolute path payload
	absPath, _ := filepath.Abs("some/sensitive/file")

	req := &ExecutionRequest{
		ToolName: "test-bypass",
		Arguments: map[string]interface{}{
			"file": absPath,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	result, err := localTool.Execute(context.Background(), req)

	// Verify that the vulnerability is fixed
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "absolute path detected")
	}
	assert.Nil(t, result)
}
