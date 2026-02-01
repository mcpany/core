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
)

func TestLocalCommandTool_InconsistentQuoting_Python(t *testing.T) {
	// Define a tool that uses the parameter twice: once quoted, once unquoted.
	// This tricks the security check into thinking it's safe (quoted),
	// while the second usage executes it unquoted.
	tool := v1.Tool_builder{
		Name: proto.String("inconsistent-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	// Template: print('{{msg}}'); print({{msg}})
	// 1. {{msg}} inside quotes -> quoteLevel 2 detected.
	// 2. quoteLevel 2 allows " and ( and )
	// 3. User injects: __import__("os").system("echo INJECTED")
	// 4. Result: print('__import__("os").system("echo INJECTED")'); print(__import__("os").system("echo INJECTED"))
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{msg}}'); print({{msg}})"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	payload := "__import__(\"os\").system(\"echo INJECTED\")"

	inputs := map[string]interface{}{
		"msg": payload,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "inconsistent-tool",
		ToolInputs: json.RawMessage(inputBytes),
	}

	_, err := localTool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Blocked as expected. Error: %v", err)
		if strings.Contains(err.Error(), "injection detected") {
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}
	t.Fatal("Inconsistent quoting injection was NOT blocked!")
}
