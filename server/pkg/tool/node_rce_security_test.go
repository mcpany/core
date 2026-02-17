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

func TestNodeFunctionInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("node")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("code")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	// Unquoted context test
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-e", "console.log({{code}})"})
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

	// Attack payload: Function("return process")().mainModule.require("child_process").execSync("ls")
	// Resulting Node code: console.log(Function("return process")().mainModule.require("child_process").execSync("ls"))

	// We need to escape quotes for JSON
	jsonPayload := `Function(\"return process\")().mainModule.require(\"child_process\").execSync(\"echo INJECTED\")`
	jsonInput := fmt.Sprintf(`{"code": "%s"}`, jsonPayload)

	req := &ExecutionRequest{
		ToolName:   "node_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	// We expect an error containing "injection detected"
	assert.ErrorContains(t, err, "injection detected", "Expected injection detection error")
}
