// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_AwkFileWrite_Repro(t *testing.T) {
	// Define a tool that uses 'awk', which is in isShellCommand list.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool-write"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-write")

	// We attempt to pass an awk script that writes to a file
	// BEGIN { print "pwned" > "pwned.txt" }
	// This relies on awk's redirection functionality.

	tempFile := "pwned.txt"
	defer os.Remove(tempFile)

	payload := `BEGIN { print "pwned" > "` + tempFile + `" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool-write",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a vulnerable system, this Execute call will SUCCEED (or fail with execution error if awk fails, but NOT security error).
	// We want to verify that the security check BLOCKS it.

	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, the protection failed.
	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk file write payload %q was allowed!", payload)
		// Check if file was created
		if _, err := os.Stat(tempFile); err == nil {
			t.Logf("File %s was successfully created!", tempFile)
		}
		t.Fail()
	} else {
		// If error occurred, check if it's the security block
		t.Logf("Blocked with error: %v", err)
	}
}
