// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestJqInjectionSingleQuote(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("jq"),
	}.Build()

	// Simulate jq -n '"{{input}}"'
	// The filter is surrounded by SINGLE quotes in shell.
	// jq -n '"{{input}}"'
	// Wait, if I pass args via CommandLineCallDefinition Args, they are arguments.
	// Arg: '"{{input}}"' (quoted string in jq).
	// Shell will wrap this arg in quotes (e.g. single quotes).
	// So shell cmd: jq -n '"{{input}}"' (if unquoted in executor? No, executor quotes args).
	// If I use `NewLocalCommandTool`, it uses `executor.Execute`.
	// `executor.Execute` passes args to `exec.Command`.
	// `exec.Command` handles argument passing directly (no shell involved usually, unless it executes sh -c).
	// BUT `checkForShellInjection` assumes shell execution context if `isShell` or `isInterpreter`?
	// `isInterpreter` includes `jq`.

	// If `jq` is executed directly, `checkForShellInjection` logic about "Quote Level" is about the TEMPLATE string provided in `Args`.
	// If `Args` contains `"{{input}}"`, `analyzeQuoteContext` sees Level 1 (Double Quoted) because of `"`.
	// So `quoteLevel` = 1.

	// If I want Quote Level 2 (Single), I need the template to be `'{{input}}'`.
	// But `jq` strings use double quotes.
	// If I use `'{{input}}'` in jq, it's not a string, it's invalid (unless it's a filter that allows single quotes? jq doesn't).

	// So `jq` filters usually involve double quotes.
	// If the template is `"{{input}}"`, Level is 1.
	// Level 1 blocks `\`.
	// So `TestJqInjection` (which used double quotes) passed because Level 1 blocked `\`.

	// Is there a way to have `jq` input in Level 2 (Single Quote) or Level 0 (Unquoted)?
	// If template is `{{input}}` (Unquoted).
	// Then `val` is injected directly.
	// If `val` = `123`, it's a number.
	// If `val` = `"string"`, it's a string.
	// If `val` = `(env.USER)`, it might access env? No, `env` is an object.

	// If `jq` allows running arbitrary code via unquoted input?
	// `jq -n 'env'` -> outputs environment.
	// If template is `{{input}}`.
	// Input: `env`.
	// `checkUnquotedInjection` blocks `v`? No.
	// It blocks `;|&$...`
	// It does NOT block `env`.

	// So if template is unquoted, I can inject `env`.

	// Let's verify if `jq` is vulnerable if unquoted.

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-n", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("jq_unquoted"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	// Payload: env
	payload := "env"

	req := &ExecutionRequest{
		ToolName:   "jq_unquoted",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// If this passes, we are allowing `jq -n env`, which leaks environment.
	if err == nil {
		t.Log("VULNERABILITY: Validation passed for jq unquoted injection (env leakage)")
		t.Fail()
	} else {
		t.Logf("Got error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
