// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestShellInjection_BackslashInSingleQuotesBug(t *testing.T) {
	// Bug description:
	// The shell injection detector incorrectly handles backslashes inside single-quoted strings.
	// In most shells (sh, bash, etc.), single quotes are "strong quotes" where backslash is NOT an escape character.
	// However, the detector treats backslash as an escape character even inside single quotes.
	//
	// Scenario:
	// Template: echo 'foo\' {{input}}
	// The detector thinks the first quote is NOT closed because it's followed by a backslash (escaped).
	// So it thinks {{input}} is inside single quotes.
	//
	// In reality (shell):
	// 'foo\' -> string "foo\"
	// The quote IS closed.
	// So {{input}} is UNQUOTED.
	//
	// If input is "; echo pwned", the detector allows it (thinking it's quoted safe).
	// But the shell executes it as a new command.

	cmd := "bash"
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	// "echo 'foo\\' {{input}}" -> In Go string literal, double backslash means single backslash.
	// So the arg string is: echo 'foo\' {{input}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'foo\\' {{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	req := &ExecutionRequest{
		ToolName: "test",
		// Input contains a shell metacharacter ";" which should be blocked if unquoted.
		ToolInputs: []byte(`{"input": "; echo pwned"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// We expect an error because ";" is dangerous in unquoted context.
	// If the bug exists, this will verify that NO error is returned (or a different error).
	// But to prove the bug, we assert that we SHOULD get an error, and if we don't, the test fails.
	// However, to make it a "failing test before fix", I expect assert.Error to fail (i.e., err is nil).

	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Logf("Bug reproduced: No error returned for dangerous input")
	}

	assert.Error(t, err, "Should detect shell injection for unquoted input following 'escaped' quote")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
}
