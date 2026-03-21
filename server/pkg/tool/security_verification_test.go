// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_SSRF_Vulnerability(t *testing.T) {
	// 1. Create a secret file
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(secretFile, []byte("CONFIDENTIAL_DATA"), 0600)
	assert.NoError(t, err)

	// 2. Configure a LocalCommandTool that uses curl
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
		Name:        proto.String("curl-tool"),
		Description: proto.String("A curl tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()

	// Arg: curl -s {{url}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-s", "{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "curl-call-id")

	// 3. Execute with file:// URL
	req := &ExecutionRequest{
		ToolName: "curl-tool",
		Arguments: map[string]interface{}{
			"url": "file://" + secretFile,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// 4. Assert that the execution is BLOCKED
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file: scheme detected")
	assert.Contains(t, err.Error(), "local file access is not allowed")
}
