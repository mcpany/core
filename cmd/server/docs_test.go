// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGenerateMarkdown(t *testing.T) {
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"arg1": map[string]interface{}{
				"type": "string",
			},
		},
	})

	mockTool := &tool.MockTool{
		ToolFunc: func() *pb.Tool {
			return &pb.Tool{
				Name:        proto.String("test_tool"),
				Description: proto.String("A test tool"),
				InputSchema: inputSchema,
			}
		},
	}

	tools := []tool.Tool{mockTool}
	md := generateMarkdown(tools)

	assert.Contains(t, md, "## test_tool")
	assert.Contains(t, md, "A test tool")
	assert.Contains(t, md, "\"arg1\"")
}

func TestGenerateMarkdownWithAnnotations(t *testing.T) {
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"arg2": map[string]interface{}{
				"type": "integer",
			},
		},
	})

	mockTool := &tool.MockTool{
		ToolFunc: func() *pb.Tool {
			return &pb.Tool{
				Name:        proto.String("annotated_tool"),
				Description: proto.String("Tool with annotations"),
				Annotations: &pb.ToolAnnotations{
					InputSchema: inputSchema,
				},
			}
		},
	}

	tools := []tool.Tool{mockTool}
	md := generateMarkdown(tools)

	assert.Contains(t, md, "## annotated_tool")
	assert.Contains(t, md, "\"arg2\"")
}

func TestDocsCmd_E2E(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()

	// Create a config file
	configContent := `
upstream_services:
  - name: "test-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "example_tool"
          description: "An example tool"
          call_id: "example_call"
      calls:
        example_call:
          endpoint_path: "/api"
          method: "HTTP_METHOD_GET"
`
	err := afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newDocsCmd(fs)
	config.BindRootFlags(cmd)
	cmd.SetArgs([]string{"--config-path", "config.yaml"})
	err = cmd.Execute()

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var out bytes.Buffer
	_, _ = io.Copy(&out, r)
	output := out.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "# Tool Documentation")
	assert.Contains(t, output, "## example_tool")
	assert.Contains(t, output, "An example tool")
}
