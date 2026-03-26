// Copyright 2025 Author(s) of MCP Any
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

func TestImageMagick_Injection_Security(t *testing.T) {
	// Represents a tool definition for ImageMagick 'convert'
	// convert {{input}} output.png
	cmd := "convert"
	toolDef := v1.Tool_builder{Name: proto.String("convert-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{input}}", "output.png"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Case 1: ImageMagick label:@/etc/passwd (LFI)
	t.Run("label_scheme", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "convert",
			ToolInputs: []byte(`{"input": "label:@/etc/passwd"}`),
			DryRun:   true,
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "dangerous scheme detected")
		}
	})

	// Case 2: ImageMagick mvg:payload.mvg (RCE/LFI)
	t.Run("mvg_scheme", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "convert",
			ToolInputs: []byte(`{"input": "mvg:payload.mvg"}`),
			DryRun:   true,
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "dangerous scheme detected")
		}
	})

	// Case 3: FFmpeg concat:file1|file2 (LFI/RCE)
	t.Run("ffmpeg_concat_scheme", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "ffmpeg",
			ToolInputs: []byte(`{"input": "concat:file1|file2"}`),
			DryRun:   true,
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "dangerous scheme detected")
		}
	})
}
