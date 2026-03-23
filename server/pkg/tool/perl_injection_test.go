// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Perl_RCE(t *testing.T) {
	// This test demonstrates RCE vulnerability in Perl using qx//
	tool := v1.Tool_builder{Name: proto.String("test-tool-perl")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
		},
		Args: []string{"-e", "{{code}}"},
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attempt RCE using qx// which uses safe characters
	// print+qx/id/
    // This payload contains NO spaces or other blocked characters.
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-perl",
		Arguments: map[string]interface{}{
			"code": "print+qx/id/",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err := localTool.Execute(context.Background(), reqAttack)

	// If err contains "shell injection detected", it means validation failed -> Secure.
	if err != nil {
        // If it fails with something else (e.g. executable not found), it means validation PASSED (Vulnerable)
        // because "executable not found" is a runtime error, not a validation error.
		if assert.Contains(t, err.Error(), "shell injection detected") {
             // Validation worked.
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Perl RCE payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}
