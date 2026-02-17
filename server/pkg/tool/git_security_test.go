// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGitSCPSecurity(t *testing.T) {
	// Mock validation.IsSafeURL to simulate production behavior (blocking non-http/https)
	// This is necessary because TestMain globally mocks it to allow everything.
	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error {
		if strings.HasPrefix(urlStr, "ssh://") {
			return assert.AnError // Simulate unsafe URL error
		}
		return nil
	}
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

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
	cmdTool := tool.NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "clone")

	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name:    "Basic SCP Injection",
			payload: "user@-oProxyCommand=touch%20/tmp/pwned:repo",
			wantErr: "git scp-style injection",
		},
		{
			name:    "Bypass Attempt (Double User)",
			payload: "fakeuser@realuser@-oProxyCommand=touch%20/tmp/pwned:repo",
			wantErr: "git scp-style injection",
		},
		{
			name:    "SSH Scheme Injection",
			payload: "ssh://-oProxyCommand=touch%20/tmp/pwned/x",
			wantErr: "unsafe url argument", // IsSafeURL blocks non-http/https
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &tool.ExecutionRequest{
				ToolName: "git_clone",
				ToolInputs: []byte(`{"url": "` + tt.payload + `"}`),
			}
			_, err := cmdTool.Execute(context.Background(), req)
			if err == nil {
				t.Fatalf("Expected error for payload %q, got nil", tt.payload)
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
