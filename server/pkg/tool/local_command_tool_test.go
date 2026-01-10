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
	tool := &v1.Tool{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("echo")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("args")}},
		},
	}

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
	tool := &v1.Tool{
		Name: proto.String("test-tool-env"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		Env: map[string]*configv1.SecretValue{
			"MY_ENV": {
				Value: &configv1.SecretValue_PlainText{
					PlainText: "secret_value",
				},
			},
		},
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "echo -n $MY_ENV"},
	}

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
	assert.Equal(t, "secret_value", resultMap["stdout"])
}

func TestLocalCommandTool_Execute_BlockedByPolicy(t *testing.T) {
	t.Parallel()
	tool := &v1.Tool{
		Name:        proto.String("test-tool-blocked"),
		Description: proto.String("A test tool"),
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("echo")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{}

	action := configv1.CallPolicy_DENY
	policies := []*configv1.CallPolicy{
		{
			DefaultAction: &action,
		},
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
	tool := &v1.Tool{
		Name:        proto.String("test-tool-json-stderr"),
		Description: proto.String("A test tool that fails"),
	}
	// Command that writes to stderr and exits with error, producing invalid JSON (empty stdout)
	service := &configv1.CommandLineUpstreamService{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "echo 'something went wrong' >&2; exit 1"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-json-stderr",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)

	// The error should NOT contain the stderr content anymore
	assert.NotContains(t, err.Error(), "something went wrong")
	// But it should suggest checking the logs
	assert.Contains(t, err.Error(), "Check logs for details")
}
