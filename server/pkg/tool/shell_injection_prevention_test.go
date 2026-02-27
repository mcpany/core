// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestShellInjectionPrevention_NestedQuotes(t *testing.T) {
	// Verify protection against SQL Injection via Shell Script nested quotes.
	// This ensures that single quotes are blocked inside double-quoted shell arguments,
	// preventing breakouts from inner contexts (like SQL strings inside shell scripts).

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	tool := v1.Tool_builder{
		Name:        proto.String("shell-sqlite-tool"),
		Description: proto.String("A wrapped sqlite tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	// bash -c "echo start; sqlite3 test.db \"SELECT * FROM users WHERE name = '{{name}}'\""
	// Double-quoted shell argument containing single-quoted SQL string.
	// We use "echo start; ..." to evade checkArgumentInterpreterInjection which only checks the first word.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo start; sqlite3 test.db \"SELECT * FROM users WHERE name = '{{name}}'\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("name")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload containing single quote
	req := &ExecutionRequest{
		ToolName: "shell-sqlite-tool",
		ToolInputs: []byte(`{"name": "admin' OR 1=1 --"}`),
		Arguments: map[string]interface{}{
			"name": "admin' OR 1=1 --",
		},
	}

	_, err := localTool.Execute(context.Background(), req)

	// We expect an error blocking the single quote
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "single quote inside double-quoted shell argument")
}
