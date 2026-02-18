// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGitSSHInjection(t *testing.T) {
	// Setup the tool configuration using builders
	cloneCall := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "target_dir"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("git_clone"),
				CallId: proto.String("clone"),
			}.Build(),
		},
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"clone": cloneCall,
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name: proto.String("git_clone"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": {
					Kind: &structpb.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"url": {
									Kind: &structpb.Value_StructValue{
										StructValue: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"type": {Kind: &structpb.Value_StringValue{StringValue: "string"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}.Build()

	// Create the tool
	tool := NewLocalCommandTool(toolDef, serviceConfig, serviceConfig.GetCalls()["clone"], nil, "clone")

	// Payload that attempts to inject SSH options via the URL using ssh: scheme (no //)
	// This currently bypasses IsSafeURL check in types.go
	payload := "ssh:-oProxyCommand=touch%20/tmp/pwned/x"

	req := &ExecutionRequest{
		ToolName: "git_clone",
		ToolInputs: []byte(`{"url": "` + payload + `"}`),
	}

	// Execute the tool
	_, err := tool.Execute(context.Background(), req)

	// Currently this should be nil (allowed) because we haven't fixed it yet.
	// We want to asserting that we fix this.
	// So IF err is nil, it means Vulnerability EXISTS.

	if err == nil {
		t.Log("Vulnerability confirmed: ssh:-o... was allowed")
        // We fail the test to signal that we need to fix it.
        // Once fixed, we expect err != nil.
        t.Fatal("Expected error due to dangerous URL, but got nil")
	} else {
        t.Logf("Blocked: %v", err)
    }
}
