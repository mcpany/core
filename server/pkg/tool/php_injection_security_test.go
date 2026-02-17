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
	"google.golang.org/protobuf/proto"
)

func TestPHPInjectionRepro(t *testing.T) {
	// Define a PHP command tool that executes code via -r
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}).Build()

	// Simulate: php -r '{{code}}'
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}).Build(),
			}).Build(),
		},
	}).Build()

	toolDef := (&v1.Tool_builder{
		Name: proto.String("php_eval"),
	}).Build()

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
