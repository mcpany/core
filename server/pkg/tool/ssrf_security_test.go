package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
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

	commandTool := tool.NewCommandTool(toolDef, svc, callDef, nil, "test-ssrf-id")

	// Payload uses gopher scheme
	payload := []byte(`{"url": "gopher://127.0.0.1:6379/_SLAVEOF..."}`)

	req := &tool.ExecutionRequest{
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

	commandTool := tool.NewCommandTool(toolDef, svc, callDef, nil, "test-ssrf-file-id")

	// Payload uses file scheme
	payload := []byte(`{"url": "file:///etc/passwd"}`)

	req := &tool.ExecutionRequest{
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
	localTool := tool.NewLocalCommandTool(toolDef, svc, callDef, nil, "test-ssrf-local-id")

	payload := []byte(`{"url": "gopher://127.0.0.1:6379/_SLAVEOF..."}`)

	req := &tool.ExecutionRequest{
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
