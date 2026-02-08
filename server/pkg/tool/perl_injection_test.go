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
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Perl_RCE(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{Name: proto.String("test-tool-perl")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Test Case 1: Unquoted (Vulnerable)
	t.Run("Unquoted qx//", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
			},
			Args: []string{"-e", "{{code}}"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-1")

		reqAttack := &ExecutionRequest{
			ToolName: "test-tool-perl",
			Arguments: map[string]interface{}{
				"code": "print+qx/id/",
			},
		}
		reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

		_, err := localTool.Execute(context.Background(), reqAttack)
		assert.Error(t, err)
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "perl qx execution detected"), "Expected qx detection, got: %v", err)
		}
	})

	// Test Case 2: Double Quoted (Vulnerable - Interpolation)
	t.Run("Double Quoted qx//", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
			},
			Args: []string{"-e", "\"{{code}}\""},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-2")

		reqAttack := &ExecutionRequest{
			ToolName: "test-tool-perl",
			Arguments: map[string]interface{}{
				"code": "print qx/id/",
			},
		}
		reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

		_, err := localTool.Execute(context.Background(), reqAttack)
		assert.Error(t, err)
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "perl qx execution detected"), "Expected qx detection, got: %v", err)
		}
	})

	// Test Case 3: Single Quoted (Safe - No Interpolation)
	t.Run("Single Quoted qx//", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
			},
			Args: []string{"-e", "'{{code}}'"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-3")

		// Input contains "qx" but in single quotes it's safe string "print qx/id/"
		reqSafe := &ExecutionRequest{
			ToolName: "test-tool-perl",
			Arguments: map[string]interface{}{
				"code": "print qx/id/",
			},
		}
		reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

		_, err := localTool.Execute(context.Background(), reqSafe)
		// Should PASS (err == nil) because it's quoted and safe.
		if err != nil {
			assert.NotContains(t, err.Error(), "perl qx execution detected")
		}
	})

	// Test Case 4: Safe Word (e.g. equinox)
	t.Run("Safe Word containing qx", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build()}.Build(),
			},
			Args: []string{"-e", "{{code}}"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-4")

		// Input contains "qx" but as part of a word "equinox"
		// Note: "equinox" uses safe chars for Unquoted context (+ - / . , : @)
		// Wait, "equinox" contains 'e', 'u', 'i', 'n', 'o'. These are letters.
		// Letters are safe in Unquoted check?
		// checkUnquotedInjection blocks: ;|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\
		// Letters are NOT blocked.

		reqSafe := &ExecutionRequest{
			ToolName: "test-tool-perl",
			Arguments: map[string]interface{}{
				"code": "print+equinox",
			},
		}
		reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

		_, err := localTool.Execute(context.Background(), reqSafe)
		// Should PASS (err == nil) because "equinox" is safe word.
		if err != nil {
			assert.NotContains(t, err.Error(), "perl qx execution detected")
		}
	})
}
