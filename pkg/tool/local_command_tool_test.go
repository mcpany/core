// Copyright 2025 Author(s) of MCP Any
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

func TestLocalCommandTool_Execute(t *testing.T) {
	tool := &v1.Tool{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("echo")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{}

	localTool := NewLocalCommandTool(tool, service, callDef)

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
	tool := &v1.Tool{
		Name:        proto.String("test-tool-env"),
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

	localTool := NewLocalCommandTool(tool, service, callDef)

	req := &ExecutionRequest{
		ToolName: "test-tool-env",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "secret_value", resultMap["stdout"])
}
