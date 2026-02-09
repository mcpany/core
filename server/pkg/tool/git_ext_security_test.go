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

func TestLocalCommandTool_GitExtExploit(t *testing.T) {
	// This test attempts to inject a git 'ext::' protocol payload.
	// We expect the security layer to BLOCK this attempt.
	// Currently, it might pass (vulnerability).

	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-git-exploit"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// Config that passes a URL to git
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "/tmp/dest"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: ext::sh -c echo pwned
	// This uses the 'ext::' protocol which allows command execution in some git versions/configurations.
	payload := "ext::sh -c echo pwned"

	req := &ExecutionRequest{
		ToolName: "test-git-exploit",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect Execute to return an error caught by the security layer.
	// If it tries to execute git, git will likely fail, but that means the security check was BYPASSED.
	_, err := localTool.Execute(context.Background(), req)

	// Assert that we caught the injection
	if err == nil {
		t.Fatal("Expected error, got nil (Exploit Successful: security check bypassed)")
	}

	// We look for a specific security error message.
	// If the error comes from 'git' (e.g. "exit status 128"), it means we FAILED to block it.
	assert.True(t, strings.Contains(err.Error(), "scheme detected") || strings.Contains(err.Error(), "dangerous"),
		"Expected security error, got: %v", err)
}
