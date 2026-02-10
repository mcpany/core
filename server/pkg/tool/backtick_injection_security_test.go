// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBacktickInjection_Repro(t *testing.T) {
	// Template: echo "`{{input}}`"
	// This puts the input inside backticks, which are inside double quotes.
	// analyzeQuoteContext should see this as Level 3 (Backtick) but likely sees Level 1 (Double).
	// Level 1 allows ';', which is dangerous inside backticks.

	cmd := "sh"
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	// We use "echo \"`{{input}}`\"" as the argument to sh -c
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo \"`{{input}}`\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	req := &ExecutionRequest{
		ToolName: "test",
		// input contains ';' which should be blocked if backtick context was correctly detected
		ToolInputs: []byte(`{"input": "true; echo vulnerable"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	// We expect this to fail with a security error.
	// If it succeeds (err == nil), it means the vulnerability is present.
	if err == nil {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout := resMap["stdout"].(string)
		if strings.Contains(stdout, "vulnerable") {
			assert.Fail(t, "Vulnerability Confirmed: Shell injection via backticks inside double quotes succeeded.")
		}
	} else {
		assert.Contains(t, err.Error(), "shell injection detected", "Should detect shell injection")
	}
}
