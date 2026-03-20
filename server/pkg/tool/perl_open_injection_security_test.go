// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPerlOpenInjection(t *testing.T) {
	// This test attempts to reproduce an RCE vulnerability where Perl open("|cmd")
	// can be injected into a SINGLE-quoted argument string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}.Build()

	// Perl script: open(F, '{{input}}'); print <F>
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "open(F, '{{input}}'); print <F>"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("perl_open_single"),
	}.Build()

	// Use LocalCommandTool to execute locally
	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: |echo PERL_RCE_SUCCESS
	payload := "|echo PERL_RCE_SUCCESS"

	req := &ExecutionRequest{
		ToolName:   "perl_open_single",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Execution blocked (good): %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)

		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "PERL_RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Perl open injection successful in single quotes")
		}
	}
}
