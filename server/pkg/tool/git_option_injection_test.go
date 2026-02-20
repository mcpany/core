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

func TestGitOptionInjection(t *testing.T) {
	// Vulnerability: Git Option Injection via SSH URL (SCP syntax)
	// If the tool uses 'git clone {{url}}', and we pass a URL like 'git@-oProxyCommand=calc:repo',
	// 'git' treats this as SSH protocol with host '-oProxyCommand=calc'.
	// This bypasses 'IsSafeURL' check because it doesn't contain '://'.
	// And it bypasses 'checkForArgumentInjection' because it doesn't start with '-'.

	t.Run("git_scp_syntax_injection", func(t *testing.T) {
		cmd := "git"
		toolDef := v1.Tool_builder{Name: proto.String("test-git")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"clone", "{{url}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		// Payload: SCP-like syntax with option injection in hostname
		// Note: We need a valid-looking SCP path: user@host:path
		// Host is -oProxyCommand=calc
		payload := "git@-oProxyCommand=calc:repo"

		req := &ExecutionRequest{
			ToolName: "test-git",
			ToolInputs: []byte(`{"url": "` + payload + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// If vulnerable, err will be nil (execution proceeded)
		if err == nil {
			t.Fatal("Expected error, got nil (VULNERABLE)")
		}

		// We expect this to fail with a security error
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "dangerous SCP-style host detected")
		}
	})

	t.Run("git_scp_syntax_bypass_attempt", func(t *testing.T) {
		cmd := "git"
		toolDef := v1.Tool_builder{Name: proto.String("test-git")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"clone", "{{url}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		// Payload: SCP-like syntax containing :// to try to bypass the check
		// If the code skips SCP check because of ://, this might pass if IsSafeURL accepts it (which it shouldn't)
		// But regardless, we want to ensure it is BLOCKED.
		payload := "git@-oProxyCommand=calc:repo/ignored://"

		req := &ExecutionRequest{
			ToolName: "test-git",
			ToolInputs: []byte(`{"url": "` + payload + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		if err != nil {
			// It might be blocked by IsSafeURL ("unsafe url argument") or SCP check ("dangerous SCP-style")
			// We just want it blocked.
			t.Logf("Error: %v", err)
		}
	})
}
