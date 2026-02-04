package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestCommandTool_InterpreterInjection_Python_SingleQuote(t *testing.T) {
	// Scenario: Config uses python -c "print('{{name}}')"
	// This uses Double Quotes for Shell.
	// Input contains ' to break out of Python string.
	// This should be blocked because it's detected as Single Quoted arg (inside python) which blocks '.

	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("python3")

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-c", "print('{{name}}')"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("greet")

	commandTool := tool.NewCommandTool(toolDef, svc, callDef, nil, "test-call-id")

	// Payload that attempts to inject code: '); import os; print("INJECTED"); print('
	payload := []byte(`{"name": "'); import os; print(\"INJECTED\"); print('"}`)

	req := &tool.ExecutionRequest{
		ToolName:   "greet",
		ToolInputs: payload,
	}

	_, err := commandTool.Execute(context.Background(), req)

	assert.Error(t, err, "Should detect interpreter injection")
	if err != nil {
		assert.Contains(t, err.Error(), "injection detected")
	}
}

func TestCommandTool_InterpreterInjection_Python_DoubleQuote(t *testing.T) {
	// Scenario: Config uses python -c 'print("{{name}}")'
	// This uses Single Quotes for Shell.
	// Input contains " to break out of Python string.
	// This should be blocked because it's detected as Double Quoted arg (inside python) which blocks ".

	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("python3")

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-c", "print(\"{{name}}\")"})

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("greet_double")

	commandTool := tool.NewCommandTool(toolDef, svc, callDef, nil, "test-call-id")

	// Payload that attempts to inject code: "); import os; print('INJECTED'); print("
	payload := []byte(`{"name": "\"); import os; print('INJECTED'); print(\""}`)

	req := &tool.ExecutionRequest{
		ToolName:   "greet_double",
		ToolInputs: payload,
	}

	_, err := commandTool.Execute(context.Background(), req)

	assert.Error(t, err, "Should detect interpreter injection")
	if err != nil {
		assert.Contains(t, err.Error(), "injection detected")
	}
}
