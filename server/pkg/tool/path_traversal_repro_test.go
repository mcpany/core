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

func TestLocalCommandTool_PathTraversal_EncodedBypass(t *testing.T) {
	t.Parallel()
	// Define a simple tool that cats a file
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"filename": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("cat-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{filename}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("filename")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attack Vector: Use encoded slash to bypass path traversal check
	// ..%2fetc%2fpasswd -> ../etc/passwd (if decoded by tool)
	// Current checkForPathTraversal only checks for literal ".." and encoded "%2e%2e" (..)
	// It does NOT check for mixed encoding or encoded slashes combined with dots.

	// We use "file" that exists to avoid unrelated errors, but the path traversal check should run first.
	attackPayload := "..%2fetc%2fpasswd"

	req := &ExecutionRequest{
		ToolName: "cat-tool",
		Arguments: map[string]interface{}{
			"filename": attackPayload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// FIXED BEHAVIOR: We expect this to FAIL with "path traversal attempt detected".
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "path traversal attempt detected", "Should detect path traversal (decoded)")
	}
}

func TestLocalCommandTool_FileScheme_EncodedBypass(t *testing.T) {
	t.Parallel()
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
		Name:        proto.String("fetcher-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("my-fetcher"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attack Vector: Use encoded 'f' to bypass "file:" check
	// %66ile:///etc/passwd -> file:///etc/passwd
	attackPayload := "%66ile:///etc/passwd"

	req := &ExecutionRequest{
		ToolName: "curl-tool",
		Arguments: map[string]interface{}{
			"url": attackPayload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// EXPECTATION: Should FAIL with "file: scheme detected" (decoded)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "file: scheme detected", "Should detect file scheme (decoded)")
	}
}
