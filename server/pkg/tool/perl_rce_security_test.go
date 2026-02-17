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

func TestPerlReadpipeInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("perl")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	// Unquoted context in Perl script
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

	// Attack payload: readpipe ls
    // This executes readpipe("ls") in Perl (assuming strict is off, which is default for -e)
    // readpipe executes the command and returns output.
	payload := "readpipe ls"

	jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "perl_rce",
		ToolInputs: []byte(jsonInput),
	}

	// We expect this to fail with "injection detected".
	_, err := tool.Execute(ctx, req)

	if err == nil {
		// Vulnerability confirmed: readpipe was NOT blocked!
        assert.Error(t, err, "Expected error for readpipe injection")
	} else {
		// If error occurred, check if it is the injection detection error
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
