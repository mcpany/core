// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestPHPInjectionRepro(t *testing.T) {
	// Define a PHP command tool that executes code via -r
	// Using setters to ensure compatibility with opaque structs
	service := &configv1.CommandLineUpstreamService{}
	service.SetCommand("php")

	// Simulate: php -r '{{code}}'
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-r", "'{{code}}'"})

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("code")

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolDef := &v1.Tool{}
	toolDef.SetName("php_eval")

	cmdTool := tool.NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload using passthru
	// Since the template is '{{code}}', we are in single quotes.
	// Single quotes allow most characters except single quotes.
	// checkInterpreterFunctionCalls should block dangerous functions.
	// But passthru is missing from the list.
	payload := []byte(`{"code": "passthru(\"echo vulnerable\");"}`)

	req := &tool.ExecutionRequest{
		ToolName:   "php_eval",
		ToolInputs: payload,
		DryRun:     true,
	}

	_, err := cmdTool.Execute(context.Background(), req)

	// We expect an error catching the injection
	// If err is nil, it means the vulnerability is reproduced (injection allowed)
	if err == nil {
		t.Fatal("Vulnerability reproduced: passthru injection was not blocked")
	}

	assert.Error(t, err, "Expected validation error for passthru injection")
	if err != nil {
		assert.Contains(t, err.Error(), "interpreter injection detected", "Error message should mention injection detection")
	}
}
