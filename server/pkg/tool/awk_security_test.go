// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_AwkInjection_Repro(t *testing.T) {
	// Check if awk is available
	awkPath, err := exec.LookPath("awk")
	if err != nil {
		t.Skip("awk not found in path")
	}

	// Define a tool that uses 'awk'.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"file": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"type": structpb.NewStringValue("string"),
						},
					}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("awk-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String(awkPath), // Use absolute path to ensure execution
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk -f {{file}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-f", "{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Create local pwn file in current directory to bypass absolute path check in LocalCommandTool
	localPwn := "pwn_repro.awk"
	err = os.WriteFile(localPwn, []byte(`BEGIN { system("echo pwned") }`), 0644)
	if err != nil {
		t.Fatalf("Failed to create local pwn.awk: %v", err)
	}
	defer func() {
		_ = os.Remove(localPwn)
	}()

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"file": localPwn,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// If vulnerable (no sandbox), awk executes system("echo pwned") -> stdout contains "pwned".
	// If fixed (sandbox), awk fails or prints error.

	if err != nil {
		t.Logf("Execute returned error (expected with sandbox): %v", err)
	} else {
		resMap := result.(map[string]interface{})
		stdout := resMap["stdout"].(string)
		stderr := resMap["stderr"].(string)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)

		// Verification: Ensure "pwned" is NOT in stdout
		assert.NotContains(t, stdout, "pwned", "Sandbox should prevent execution of system command")

		// Verification: Ensure we see sandbox error in stderr
		assert.Contains(t, stderr, "sandbox mode", "Stderr should mention sandbox mode violation")
	}
}
