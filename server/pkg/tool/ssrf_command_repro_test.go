// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_SSRF_Protection(t *testing.T) {
	// This test verifies that we block arguments that look like internal URLs or IPs.
	// We use "echo" as the command, but the validation logic runs before execution.

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"url": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-ssrf"),
		Description: proto.String("A test tool for SSRF"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build()}.Build(),
		},
		Args: []string{"{{url}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	tests := []struct {
		name      string
		input     string
		shouldFail bool // After fix, this should be true for internal IPs
	}{
		{
			name:      "Loopback URL",
			input:     "http://127.0.0.1/admin",
			shouldFail: true,
		},
		{
			name:      "Loopback IP",
			input:     "127.0.0.1",
			shouldFail: true,
		},
		{
			name:      "Localhost URL",
			input:     "http://localhost:8080",
			shouldFail: true,
		},
		{
			name:      "Private IP URL",
			input:     "http://192.168.1.1",
			shouldFail: true,
		},
		{
			name:      "Metadata Service",
			input:     "http://169.254.169.254",
			shouldFail: true,
		},
		{
			name:      "Public IP",
			input:     "8.8.8.8",
			shouldFail: false,
		},
		{
			name:      "Public URL",
			input:     "http://example.com",
			shouldFail: false,
		},
		{
			name:      "Normal String",
			input:     "hello world",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "test-tool-ssrf",
				Arguments: map[string]interface{}{
					"url": tt.input,
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			result, err := localTool.Execute(context.Background(), req)

			// BEFORE FIX: Expecting failure only if existing validation blocks it (e.g. file: scheme)
			// But for http/ip, it currently allows.
			// To demonstrate repro, we assert that it SUCCEEDS for now (proving vulnerability),
			// or we check if err is nil.

			if tt.shouldFail {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				if err != nil {
					assert.Contains(t, err.Error(), "SSRF attempt blocked")
				}
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
			}

			// Clean up result for next step (Wait, I can't modify the code between runs in a single tool call)
			_ = result
		})
	}
}
