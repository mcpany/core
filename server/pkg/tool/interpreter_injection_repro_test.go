// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_InterpreterInjection_DoubleQuoteBypass(t *testing.T) {
	t.Parallel()
	// Define a tool that uses 'python3', which is an interpreter.
	// We use double quotes in the args, which corresponds to QuoteLevel 1.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("python-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `python3 -c "{{code}}"`
	// The quotes are part of the argument string passed to shell (if shell were used)
	// or part of the template analysis.
	// Since LocalCommandTool calls checkForShellInjection, it analyzes the template.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "\"{{code}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload that uses dangerous interpreter functions but avoids characters blocked in Double Quotes (", $, `, \, %)
	// __import__('os').system('id')
	// This payload has no double quotes, no backticks, no backslashes, no $.
	payload := "__import__('os').system('id')"

	req := &ExecutionRequest{
		ToolName: "python-tool",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this to BE BLOCKED by the security checks.
	_, err := localTool.Execute(context.Background(), req)

	// Expect failure due to interpreter injection detection
	if err == nil {
		t.Fatal("Expected error due to interpreter injection, but got nil")
	}

	// The error message might vary slightly but should indicate injection detection
	if !strings.Contains(err.Error(), "injection detected") {
		t.Fatalf("Expected 'injection detected' error, but got: %v", err)
	}
}
