// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGitSCPSecurity(t *testing.T) {
	// Create the call definition using builder
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "target_dir"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{
					Name: proto.String("url"),
					Type: configv1.ParameterType_STRING.Enum(),
				}).Build(),
			}).Build(),
		},
	}).Build()

	// Create the service configuration
	serviceConfig := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Tools: []*configv1.ToolDefinition{
			(&configv1.ToolDefinition_builder{
				Name:   proto.String("git_clone"),
				CallId: proto.String("clone"),
			}).Build(),
		},
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"clone": callDef,
		},
	}).Build()

	// Create the tool definition
	toolDef := (&v1.Tool_builder{
		Name: proto.String("git_clone"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"url": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"type": structpb.NewStringValue("string"),
							},
						}),
					},
				}),
			},
		},
	}).Build()

	// Create the tool
	// We use NewLocalCommandTool because we want to test local execution path validation
	cmdTool := tool.NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "clone")

	// Payload that attempts to inject SSH options via SCP-like syntax
	// git clone user@-oProxyCommand=touch%20/tmp/pwned:repo
	// This makes git use ssh with -oProxyCommand=... as the hostname, which ssh interprets as an option.
	payload := "user@-oProxyCommand=touch%20/tmp/pwned:repo"

	req := &tool.ExecutionRequest{
		ToolName: "git_clone",
		ToolInputs: []byte(`{"url": "` + payload + `"}`),
	}

	// Execute the tool
	_, err := cmdTool.Execute(context.Background(), req)

	// We expect validation error.
	if err == nil {
		t.Fatal("Expected error due to dangerous SCP-like URL, but got nil")
	}

	// Check if the error is what we expect (once implemented)
	// For now, it might fail with "exit status 128" (git error) or pass if validation is missing.
	// If validation is missing, git might try to clone and fail.
	// But we WANT it to fail validation BEFORE execution.

	t.Logf("Got error: %v", err)
	assert.ErrorContains(t, err, "git scp-style injection")
}
