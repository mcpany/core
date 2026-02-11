// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
    "path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_FileAccessBypass_Vulnerability(t *testing.T) {
	// 1. Create a secret file
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(secretFile, []byte("CONFIDENTIAL_DATA"), 0600)
	assert.NoError(t, err)

	tool := v1.Tool_builder{
		Name: proto.String("bypass-tool"),
	}.Build()

    // Configure service with BOTH Local=true AND ContainerEnvironment
    // This triggers the logic bug where isDocker=true (bypassing checks)
    // but execution is still local.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
        ContainerEnvironment: configv1.ContainerEnvironment_builder{
            Image: proto.String("dummy-image"), // Needed for isDocker check
        }.Build(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{path}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("path"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "bypass-call-id")

	// 2. Execute with absolute path to secret file
    // Normally this is blocked by checkForLocalFileAccess
	req := &ExecutionRequest{
		ToolName: "bypass-tool",
		Arguments: map[string]interface{}{
			"path": secretFile,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

    // 3. Verify FIX
    // Execution should now be BLOCKED.
    if err != nil {
        t.Logf("Execution blocked as expected: %v", err)
        assert.Error(t, err)
        assert.Nil(t, result)
        // Verify the error message indicates local file access is blocked
        // It might be "absolute path detected" or "local file access is not allowed" depending on input
        if filepath.IsAbs(secretFile) {
            assert.Contains(t, err.Error(), "absolute path detected")
        } else {
            assert.Contains(t, err.Error(), "local file access is not allowed")
        }
    } else {
        t.Errorf("VULNERABILITY CONFIRMED: Local file access bypass via dummy container config. Execution succeeded unexpectedly.")
        if resultMap, ok := result.(map[string]interface{}); ok {
             if output, ok := resultMap["combined_output"].(string); ok {
                 t.Logf("Output: %s", output)
             }
        }
    }
}
