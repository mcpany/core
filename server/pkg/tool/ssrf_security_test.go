// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF_Prevention(t *testing.T) {
	// Not Parallel because we modify global variable validation.IsSafeURL

	// Mock IsSafeURL to fail for a specific unsafe URL, simulating the real behavior
	// without breaking other tests that might use localhost.
	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error {
		// Mimic IsSafeURL logic for unsupported schemes (which IsSafeURL does)
		// But here we rely on the REAL IsSafeURL logic being called in production.
		// Since we mocked it, we must replicate the behavior we want to test:
		// 1. It is CALLED.
		// 2. It receives the correct string.

		if urlStr == "http://unsafe.local" {
			return fmt.Errorf("unsafe url detected")
		}
		if urlStr == "HTTP://unsafe.local" {
			return fmt.Errorf("unsafe url detected (case insensitive)")
		}
		// Mimic IsSafeURL blocking non-http
		if urlStr == "ftp://example.com" {
			return fmt.Errorf("unsupported scheme: ftp")
		}

		return nil
	}
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-curl"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build()}.Build(),
		},
		Args: []string{"{{url}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	testCases := []struct {
		name      string
		input     string
		shouldErr bool
		errContains string
	}{
		{"Unsafe URL", "http://unsafe.local", true, "unsafe url"},
		{"Unsafe URL Uppercase", "HTTP://unsafe.local", true, "unsafe url"},
		{"Unsupported Scheme", "ftp://example.com", true, "unsafe url"}, // Checks error wrapping
		{"Safe URL", "http://example.com", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "test-tool-curl",
				Arguments: map[string]interface{}{
					"url": tc.input,
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			_, err := localTool.Execute(context.Background(), req)

			if tc.shouldErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				// We expect NO validation error. Execution error (command not found) is fine.
				if err != nil {
					assert.NotContains(t, err.Error(), "unsafe url")
				}
			}
		})
	}
}
