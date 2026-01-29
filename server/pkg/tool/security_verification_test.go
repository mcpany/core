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

func TestLocalCommandTool_RCE_Perl_Blocked(t *testing.T) {
	// Setup
	tmpFile := "/tmp/rce_pwned"
	_ = os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("rce-tool"),
		Description: proto.String("A vulnerable tool"),
		InputSchema: inputSchema,
	}.Build()

	// Configure a service that uses Perl
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// The args use a placeholder {{code}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "rce-call-id")

	// The payload uses qx/.../ to bypass blocked characters like quotes and backticks
	// We want to execute: touch /tmp/rce_pwned
	// We use comma as delimiter because / is used in the path and would need escaping,
	// and we can't escape because backslash is blocked.
	payload := "print qx,touch " + tmpFile + ","

	req := &ExecutionRequest{
		ToolName: "rce-tool",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// Expect error due to strict security check
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
		assert.Contains(t, err.Error(), ",")
	}

	// Verify if the file was created (should NOT be created)
	_, statErr := os.Stat(tmpFile)
	if !os.IsNotExist(statErr) {
		t.Fatalf("RCE Successful! File %s was created.", tmpFile)
	}
}

func TestLocalCommandTool_Curl_Comma_Allowed(t *testing.T) {
	// This test ensures that legitimate tools like curl can use commas in arguments
	// because they are removed from the strict shell check list.

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
		Description: proto.String("Curl tool"),
		InputSchema: inputSchema,
	}.Build()

	// Configure a service that uses Curl
	// We check for validation, so even if execution fails (curl not found), validation should pass.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-I", "{{url}}"}, // Head request
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "curl-call-id")

	// URL with comma
	payload := "http://example.com/foo,bar"

	req := &ExecutionRequest{
		ToolName: "curl-tool",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// Should NOT return shell injection error
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected")
	}
}
