// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSSRFLoopbackShorthand(t *testing.T) {
	// Ensure loopback is disallowed
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// Unset dangerous bypass to ensure IsSafeIP works correctly
	originalDangerous := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	defer func() {
		if originalDangerous != "" {
			os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalDangerous)
		}
	}()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("arg"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name: proto.String("test_tool"),
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test_call")

	tests := []struct {
		name      string
		input     string
		shouldErr bool
		errorMsg  string
	}{
		{"Standard Loopback", "127.0.0.1", true, "loopback address is not allowed"},
		{"Loopback Shorthand 127.1", "127.1", true, "loopback shorthand is not allowed"},
		{"Loopback Shorthand 127.0.1", "127.0.1", true, "loopback shorthand is not allowed"},
		{"Loopback Shorthand 127.255", "127.255", true, "loopback shorthand is not allowed"},
		{"Loopback Shorthand 127.1.1", "127.1.1", true, "loopback shorthand is not allowed"},
		{"Valid Number", "10", false, ""},
		{"Valid Version", "10.1", false, ""}, // net.ParseIP fails, IsLoopbackShorthand false
		{"Zero", "0", false, ""},             // net.ParseIP fails, IsLoopbackShorthand false
		{"Filename", "127.txt", false, ""},
		{"Filename with digits", "127.1.txt", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName:   "test_tool",
				ToolInputs: []byte(`{"arg": "` + tt.input + `"}`),
			}
			_, err := tool.Execute(context.Background(), req)
			if tt.shouldErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
