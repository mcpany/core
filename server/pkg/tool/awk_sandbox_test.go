// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"google.golang.org/protobuf/proto"
)

func TestAwkFileRead_Repro(t *testing.T) {
	// 1. Create a secret file
	tmpDir, err := os.MkdirTemp("", "awk-repro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Allow tmpDir
	validation.SetAllowedPaths([]string{tmpDir})
	defer validation.SetAllowedPaths(nil)

	secretFile := filepath.Join(tmpDir, "secret.txt")
	secretContent := "SUPER_SECRET_TOKEN"
	if err := os.WriteFile(secretFile, []byte(secretContent), 0600); err != nil {
		t.Fatal(err)
	}

	// 2. Create a malicious awk script in the SAME directory
	scriptFile := "exploit.awk"
	// Note: We use absolute path for secret file in the script
	scriptContent := "BEGIN { while ((getline line < \"" + secretFile + "\") > 0) print line; close(\"" + secretFile + "\") }"
	if err := os.WriteFile(filepath.Join(tmpDir, scriptFile), []byte(scriptContent), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := "awk"
	// Tool: awk -f {{script}}
	toolDef := (&v1.Tool_builder{Name: proto.String("awk-tool")}).Build()
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
		WorkingDirectory: proto.String(tmpDir), // Run in temp dir
	}).Build()
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-f", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{Name: proto.String("script")}).Build(),
			}).Build(),
		},
	}).Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: Relative path to the script
	inputs := map[string]string{
		"script": scriptFile,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		// If execution fails, it might be due to sandbox error (good)
		t.Logf("Execute failed (expected?): %v", err)
		return
	}

	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not map: %v", result)
	}

	stdout, _ := resMap["stdout"].(string)
	t.Logf("Stdout: %s", stdout)

	if strings.Contains(stdout, secretContent) {
		t.Errorf("FAIL: Arbitrary File Read via awk succeeded! Sandbox failed.")
	} else {
		t.Logf("Success: Secret not found in output. Stderr: %v", resMap["stderr"])
	}
}

func TestAwkGetlineInjection(t *testing.T) {
	cmd := "awk"
	// Tool: awk '{{script}}'
	// Intent: Inline script.

	toolDef := (&v1.Tool_builder{Name: proto.String("awk-inline")}).Build()
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}).Build()
	// Use single quotes to trigger quoteLevel 2 check
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{Name: proto.String("script")}).Build(),
			}).Build(),
		},
	}).Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: BEGIN { getline ... }
	script := "BEGIN { getline line < \"/etc/passwd\"; print line }"

	inputs := map[string]string{
		"script": script,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName: "awk-inline",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)
	if err == nil {
		t.Errorf("Execute succeeded unexpectedly! Result: %v", result)
	} else {
		if strings.Contains(err.Error(), "awk injection detected") && strings.Contains(err.Error(), "getline") {
			t.Logf("Success: Blocked by input validation: %v", err)
		} else {
			// It might be blocked by sandbox if input validation failed?
			// But checkInterpreterInjection runs BEFORE execution.
			// However, if checkInterpreterInjection misses it, and sandbox catches it, error would be from Execute (stderr).
			// Here err != nil means Execute failed OR input validation failed.
			// NewLocalCommandTool returns a tool. Execute is called on tool.
			// Execute calls validateSafePathAndInjection and checkForShellInjection.
			t.Logf("Execute failed with error: %v", err)
		}
	}
}
