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

func TestNodeProcessInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("node")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("code")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-e", "{{code}}"})
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

	// Attack payload: process.mainModule['req'+'uire']('chi'+'ld_process').execSync('ls')
	payload := `process.mainModule['req'+'uire']('chi'+'ld_process').execSync('ls')`

	jsonInput := fmt.Sprintf(`{"code": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "node_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
		// Vulnerability confirmed
        assert.Error(t, err, "Expected error for node injection")
	} else {
        t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
