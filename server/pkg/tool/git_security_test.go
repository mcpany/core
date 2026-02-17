// Copyright 2026 Author(s) of MCP Any
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
	// We avoid using '/' in the host part because if it contains '/', git treats it as a local path, not an SCP URL.
	// This payload attempts to disable strict host key checking by injecting -o option.
	payload := "user@-oStrictHostKeyChecking=no:repo"

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

	t.Logf("Got error: %v", err)
	assert.ErrorContains(t, err, "git scp-style injection")

	// Bypass attempt: Injecting :// in the path to trick validation into thinking it's a URL scheme
	// payload: user@-oStrictHostKeyChecking=no:repo/with/://extra
	payloadBypass := "user@-oStrictHostKeyChecking=no:repo/with/://extra"
	reqBypass := &tool.ExecutionRequest{
		ToolName: "git_clone",
		ToolInputs: []byte(`{"url": "` + payloadBypass + `"}`),
	}

	_, errBypass := cmdTool.Execute(context.Background(), reqBypass)
	if errBypass == nil {
		t.Fatal("Expected error due to dangerous SCP-like URL with :// bypass attempt, but got nil")
	}
	t.Logf("Got bypass error: %v", errBypass)
	// It should either be caught by IsSafeURL (if it treats it as invalid scheme) or checkGitSCPInjection.
	// Since checkGitSCPInjection runs after IsSafeURL (if :// present), and we removed the skip check,
	// it should be caught by one of them.
	// If IsSafeURL is triggered, it fails because scheme is invalid or unsupported.
	// If IsSafeURL is NOT triggered (e.g. if we rely on start of string), then checkGitSCPInjection must catch it.
	// In our implementation, IsSafeURL is triggered if Contains("://").
	// So IsSafeURL runs first.
	// Scheme parsing of "user@-o...:repo/with/://..."?
	// url.Parse might fail or return empty scheme. IsSafeURL fails on empty/invalid scheme.
	// So we expect AN error.
	assert.Error(t, errBypass)
}
