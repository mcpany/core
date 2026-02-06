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

func TestLocalFileAccess_SensitiveFiles(t *testing.T) {
	cmd := "cat"
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	// No working directory specified, defaults to CWD
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build(),
			}.Build(),
		},
	}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	testCases := []struct {
		filename string
		blocked  bool
	}{
		{".env", true},
		{".env.local", true},
		{".git/config", true}, // .git directory blocked
		{"server.key", true},
		{"cert.pem", true},
		{"id_rsa", false}, // Not explicitly blocked by extension, though risky. We focus on specific exts.
		{"README.md", false},
		{"go.mod", false},
		{"subdir/.env", true},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "test",
				ToolInputs: []byte(`{"file": "` + tc.filename + `"}`),
			}

			_, err := tool.Execute(context.Background(), req)
			if tc.blocked {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "access to sensitive file", "Expected blocking for %s", tc.filename)
			} else {
				// We expect execution to proceed (it might fail because cat fails or file doesn't exist, but NOT blocked by security check)
				// The security check returns error BEFORE execution starts?
				// Actually LocalCommandTool.Execute runs validation first.
				// If validation passes, it tries to execute 'cat filename'.
				// If file doesn't exist, 'cat' returns error.
				// But we are checking for the security error specifically.
				if err != nil {
					assert.NotContains(t, err.Error(), "access to sensitive file", "Should not block %s", tc.filename)
				}
			}
		})
	}
}
