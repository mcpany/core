// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_Ext_Security(t *testing.T) {
	// This test verifies that git command execution blocks dangerous protocols like ext::

    // Setup temporary file for proof (should not be created)
    tmpFile := "/tmp/pwned_git_ext"
    os.Remove(tmpFile)
    defer os.Remove(tmpFile)

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("git-clone"),
	}.Build()

    toolProto := v1.Tool_builder{
        Name: proto.String("git-clone"),
    }.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
        Tools:   []*configv1.ToolDefinition{tool},
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "/tmp/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: ext::sh -c "touch /tmp/pwned_git_ext"
	payload := "ext::sh -c touch /tmp/pwned_git_ext"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	_, err := localTool.Execute(ctx, req)

    // Expect error
    assert.Error(t, err)
    if err != nil {
        assert.Contains(t, err.Error(), "dangerous scheme detected")
        assert.Contains(t, err.Error(), "ext::")
    }

    // Check if file was created (it should NOT be)
    if _, err := os.Stat(tmpFile); err == nil {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git ext:: protocol. File %s created.", tmpFile)
	}
}
