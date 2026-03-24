// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTarEnvInjection_Protection(t *testing.T) {
	// This test ensures that TAR_OPTIONS injection via environment variables is prevented
    // either by argument injection checks or by the tool itself rejecting invalid options.

	// Setup a dummy file to tar
	err := os.WriteFile("testfile.txt", []byte("data"), 0644)
	require.NoError(t, err)
	defer os.Remove("testfile.txt")
	defer os.Remove("archive.tar")
	defer os.Remove("pwned")

	toolDef := v1.Tool_builder{
		Name: proto.String("tar-env-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("tar"),
		Local:   proto.Bool(true),
	}.Build()

	// Tool definition allowing TAR_OPTIONS parameter
	callDefVulnerable := configv1.CommandLineCallDefinition_builder{
		Args: []string{"cf", "archive.tar", "testfile.txt"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("TAR_OPTIONS")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDefVulnerable, nil, "call-id")

	// Payload to execute "touch pwned"
	// We use "v" to try to bypass initial checks, but expect it to fail or be blocked.
	payload := "v --checkpoint=1 --checkpoint-action=exec=touch pwned"

	req := &ExecutionRequest{
		ToolName: "tar-env-tool",
		Arguments: map[string]interface{}{
			"TAR_OPTIONS": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, _ = localTool.Execute(context.Background(), req)

    // We expect failure (tar complaining) or blocking.
    // In any case, pwned must not exist.

	// Check if pwned file was created
	_, err = os.Stat("pwned")
	assert.True(t, os.IsNotExist(err), "pwned file should not be created")
}
