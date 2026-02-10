// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Perl_Open_Injection(t *testing.T) {
	// This test demonstrates that perl tools are vulnerable to magic open() behavior
    // if input contains pipes, even if correctly quoted.

    // Setup temporary file for proof
    tmpFile := "/tmp/pwned_perl_open"
    os.Remove(tmpFile)
    defer os.Remove(tmpFile)

	toolProto := v1.Tool_builder{
		Name: proto.String("perl-open-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

    // Simulate a perl script that opens a file specified by user input
    // Using double quotes for interpolation.
    // perl -e 'open(F, "{{input}}");'
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "open(F, \"{{input}}\");"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: command ending with |
    // "echo PWNED > /tmp/pwned_perl_open |"
    // Inside the script it becomes: open(F, "echo PWNED > /tmp/pwned_perl_open |");
    // which executes the command.
	payload := "echo PWNED > /tmp/pwned_perl_open |"

	req := &ExecutionRequest{
		ToolName: "perl-open-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	_, err := localTool.Execute(ctx, req)

    // We expect an error now because the input should be blocked
    if err == nil {
        t.Error("Expected error due to blocked input, but got nil")
    } else {
        t.Logf("Got expected error: %v", err)
        if !strings.Contains(err.Error(), "perl magic open injection detected") {
             t.Errorf("Expected error message to contain 'perl magic open injection detected', got: %v", err)
        }
    }

    // Check if file was created (should not be)
    if _, err := os.Stat(tmpFile); err == nil {
		t.Errorf("VULNERABILITY: Perl open() injection executed despite error? File %s created.", tmpFile)
	} else {
        t.Log("Safe: File not created.")
	}
}
