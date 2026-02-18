// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestPerlSystemInjection(t *testing.T) {
	runPerlInjectionTest(t, "system ls", "system")
}

func TestPerlReadpipeInjection(t *testing.T) {
    runPerlInjectionTest(t, "readpipe ls", "readpipe")
}

func TestPerlSyscallInjection(t *testing.T) {
    runPerlInjectionTest(t, "syscall 1", "syscall")
}

func runPerlInjectionTest(t *testing.T, payload string, expectedKeyword string) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("perl")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-e", "print {{name}}"})
	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolStruct := &v1.Tool{}
	toolStruct.SetName("perl_rce")

	tool := NewLocalCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "perl_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
		assert.Fail(t, fmt.Sprintf("%s injection was not blocked", payload))
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
        assert.Contains(t, err.Error(), expectedKeyword, "Expected detection of keyword")
	}
}

func TestNodeFunctionInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("node")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-e", "console.log({{name}})"})
	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolStruct := &v1.Tool{}
	toolStruct.SetName("node_rce")

	tool := NewLocalCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	// Attack payload: Function("return process")()
	payload := "Function('return process')()"

	// JSON escape single quotes
    // But here payload has single quotes inside JSON string.
    // `{"name": "Function('return process')()"}` is valid JSON.
	jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "node_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
		assert.Fail(t, "Function injection was not blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
        assert.Contains(t, err.Error(), "Function", "Expected detection of Function")
	}
}
