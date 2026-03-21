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

func TestLocalCommandTool_Perl_UnquotedArrayInjection(t *testing.T) {
	// Setup Perl Tool with Unquoted Argument
	tool := v1.Tool_builder{Name: proto.String("test-tool-perl-unquoted")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-e", "print {{msg}}"}, // Unquoted!
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-perl-unquoted")

	// Attack payload: @INC - Prints include paths.
	// This is allowed because checkUnquotedInjection does not block '@'.
	payload := "@INC"
	req := &ExecutionRequest{
		ToolName: "test-tool-perl-unquoted",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// If err contains "injection detected", it means validation failed -> Secure.
	if err != nil {
		t.Logf("Error: %v", err)
		if assert.Contains(t, err.Error(), "injection detected") {
             // Validation worked (Secure).
			 t.Log("Perl unquoted array injection correctly blocked.")
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Perl unquoted array injection payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}

func TestLocalCommandTool_Ruby_UnquotedVariableInjection(t *testing.T) {
	// Setup Ruby Tool with Unquoted Argument
	tool := v1.Tool_builder{Name: proto.String("test-tool-ruby-unquoted")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-e", "puts {{msg}}"}, // Unquoted
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-ruby-unquoted")

    payload := "@foo"

	req := &ExecutionRequest{
		ToolName: "test-tool-ruby-unquoted",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// If err contains "injection detected", it means validation failed -> Secure.
	if err != nil {
		t.Logf("Error: %v", err)
		if assert.Contains(t, err.Error(), "injection detected") {
             // Validation worked (Secure).
			 t.Log("Ruby unquoted variable injection correctly blocked.")
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Ruby unquoted variable injection payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}

func TestLocalCommandTool_Perl_DoubleQuoteArrayInjection(t *testing.T) {
	// Setup Perl Tool with Double Quoted Argument
	tool := v1.Tool_builder{Name: proto.String("test-tool-perl-double")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-e", "print \"{{msg}}\""}, // Double Quoted
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-perl-double")

	// Attack payload: @INC - Prints include paths in double quotes.
	// Currently, allowed because checkNodePerlPhpInjection only blocks @{...}, not @var.
	payload := "@INC"
	req := &ExecutionRequest{
		ToolName: "test-tool-perl-double",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// If err contains "injection detected", it means validation failed -> Secure.
	if err != nil {
		t.Logf("Error: %v", err)
		if assert.Contains(t, err.Error(), "injection detected") {
             // Validation worked (Secure).
			 t.Log("Perl double quoted array injection correctly blocked.")
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Perl double quoted array injection payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}

func TestLocalCommandTool_Ruby_DoubleQuoteInstanceInjection(t *testing.T) {
	// Setup Ruby Tool
	tool := v1.Tool_builder{Name: proto.String("test-tool-ruby-double")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-e", "puts \"{{msg}}\""},
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-ruby-double")

	// Attack payload: #@foo - Interpolates instance variable.
	// # is allowed in double quotes. @ is allowed in double quotes.
	payload := "#@foo"
	req := &ExecutionRequest{
		ToolName: "test-tool-ruby-double",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// If err contains "injection detected", it means validation failed -> Secure.
	if err != nil {
		t.Logf("Error: %v", err)
		if assert.Contains(t, err.Error(), "injection detected") {
             // Validation worked (Secure).
			 t.Log("Ruby instance variable injection correctly blocked.")
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Ruby instance variable injection payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}
