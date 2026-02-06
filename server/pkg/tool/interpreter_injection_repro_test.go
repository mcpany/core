// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRCERepro_Ruby(t *testing.T) {
	// Vulnerability: sh -c "ruby -e '{{code}}'" allows RCE via `system "cmd"` (no parens)
	// Current checks block `system(` but not `system "..."` in Single Quotes.

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "ruby -e '{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	// We use "sh" which is a shell
	cmdTool := newCommandTool("sh", callDef)

	// Payload: Try to execute system command
	// We print RCE_SUCCESS to stdout
	inputData := map[string]interface{}{
		"code": `print "START"; system "echo RCE_SUCCESS"; print "END"`,
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = cmdTool.Execute(context.Background(), req)

	// Expect error due to security fix
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interpreter injection detected")
	assert.Contains(t, err.Error(), "dangerous keyword")
}

func TestRCERepro_Perl(t *testing.T) {
	// Vulnerability: sh -c "perl -e '{{code}}'" allows RCE via `system "cmd"`

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "perl -e '{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	cmdTool := newCommandTool("sh", callDef)

	inputData := map[string]interface{}{
		"code": `print "START"; system "echo RCE_SUCCESS"; print "END"`,
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = cmdTool.Execute(context.Background(), req)

	// Expect error due to security fix
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interpreter injection detected")
	assert.Contains(t, err.Error(), "dangerous keyword")
}
