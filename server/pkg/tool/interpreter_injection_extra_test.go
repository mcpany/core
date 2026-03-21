// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSqlite3ShellInjection(t *testing.T) {
	// Setup sqlite3 tool
	toolProto := v1.Tool_builder{
		Name: proto.String("sqlite3"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sqlite3"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{":memory:", "{{query}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("query"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-call-sqlite")

	// Injection payload: .shell echo pwned
	// This should be blocked.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	inputs := map[string]interface{}{
		"query": ".shell echo pwned",
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "sqlite3",
		ToolInputs: inputsBytes,
	}

	_, err := tool.Execute(ctx, req)

	// Expect error: "SQL injection detected" or similar
	if err == nil {
		t.Error("VULNERABILITY: sqlite3 .shell injection was NOT blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detected error")
	}
}

func TestGdbShellInjection(t *testing.T) {
	// Setup gdb tool
	toolProto := v1.Tool_builder{
		Name: proto.String("gdb"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gdb"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-ex", "{{cmd}}", "--batch"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-call-gdb")

	// Injection payload: shell echo pwned
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	inputs := map[string]interface{}{
		"cmd": "shell echo pwned",
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "gdb",
		ToolInputs: inputsBytes,
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
		t.Error("VULNERABILITY: gdb shell injection was NOT blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detected error")
	}
}

func TestMysqlSystemInjection(t *testing.T) {
	// Setup mysql tool
	toolProto := v1.Tool_builder{
		Name: proto.String("mysql"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("mysql"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{query}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("query"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-call-mysql")

	// Injection payload: system echo pwned
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	inputs := map[string]interface{}{
		"query": "system echo pwned",
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "mysql",
		ToolInputs: inputsBytes,
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
		t.Error("VULNERABILITY: mysql system injection was NOT blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detected error")
	}
}
