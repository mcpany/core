// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestCommandTool_DangerousSchemes(t *testing.T) {
	// Scenario: Config uses curl {{url}}
	// Input uses gopher:// to attack internal services

	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("curl")

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"{{url}}"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("url")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("fetch_url")

	commandTool := NewCommandTool(toolDef, svc, callDef, nil, "test-ssrf-id")

	// Payload uses gopher scheme
	payload := []byte(`{"url": "gopher://127.0.0.1:6379/_SLAVEOF..."}`)

	req := &ExecutionRequest{
		ToolName:   "fetch_url",
		ToolInputs: payload,
	}

	_, err := commandTool.Execute(context.Background(), req)

	assert.Error(t, err, "Should detect dangerous scheme")
	if err != nil {
		assert.Contains(t, err.Error(), "dangerous scheme detected")
		assert.Contains(t, err.Error(), "gopher:")
	}
}

func TestCommandTool_DangerousSchemes_File(t *testing.T) {
	// Scenario: Config uses curl {{url}}
	// Input uses file:// to read local files
	// Environment: Host (no container env) -> Blocked

	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("curl")

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"{{url}}"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("url")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("fetch_file")

	commandTool := NewCommandTool(toolDef, svc, callDef, nil, "test-ssrf-file-id")

	// Payload uses file scheme
	payload := []byte(`{"url": "file:///etc/passwd"}`)

	req := &ExecutionRequest{
		ToolName:   "fetch_file",
		ToolInputs: payload,
	}

	_, err := commandTool.Execute(context.Background(), req)

	assert.Error(t, err, "Should detect dangerous scheme")
	if err != nil {
		assert.Contains(t, err.Error(), "dangerous scheme detected")
		assert.Contains(t, err.Error(), "file:")
	}
}

func TestCommandTool_FileScheme_AllowedInDocker(t *testing.T) {
	// Scenario: Config uses curl {{url}}
	// Input uses file:// to read local files
	// Environment: Docker -> Allowed

	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("curl")

	// Set container env to make isDocker=true
	ce := &configv1.ContainerEnvironment{}
	ce.SetImage("alpine")
	svc.SetContainerEnvironment(ce)

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"{{url}}"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("url")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("fetch_file_docker")

	commandTool := NewCommandTool(toolDef, svc, callDef, nil, "test-ssrf-docker-id")

	// Payload uses file scheme
	payload := []byte(`{"url": "file:///etc/passwd"}`)

	req := &ExecutionRequest{
		ToolName:   "fetch_file_docker",
		ToolInputs: payload,
	}

	_, err := commandTool.Execute(context.Background(), req)

	// It might fail due to execution issues (no docker), but it should NOT be "dangerous scheme detected"
	if err != nil {
		assert.NotContains(t, err.Error(), "dangerous scheme detected", "Should allow file: scheme in Docker")
	}
}

func TestLocalCommandTool_DangerousSchemes(t *testing.T) {
	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("curl")

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"{{url}}"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("url")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("fetch_local")

	// Use NewLocalCommandTool
	localTool := NewLocalCommandTool(toolDef, svc, callDef, nil, "test-ssrf-local-id")

	payload := []byte(`{"url": "gopher://127.0.0.1:6379/_SLAVEOF..."}`)

	req := &ExecutionRequest{
		ToolName:   "fetch_local",
		ToolInputs: payload,
	}

	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err, "Should detect dangerous scheme in LocalCommandTool")
	if err != nil {
		assert.Contains(t, err.Error(), "dangerous scheme detected")
		assert.Contains(t, err.Error(), "gopher:")
	}
}
