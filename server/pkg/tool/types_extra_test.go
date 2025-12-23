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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckForPathTraversal(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{"safe", false},
		{"safe/path", false},
		{"..", true},
		{"../", true},
		{"..\\", true},
		{"/..", true},
		{"\\..", true},
		{"/../", true},
		{"\\..\\", true},
		{"/..\\", true},
		{"\\../", true},
		{"foo/../bar", false}, // Safe because it resolves to "bar"
		{"foo\\..\\bar", true},
		{"../bar", true},
		{"bar/..", false}, // Safe because it resolves to "."
		{"bar\\..", true},
		{"mixed/..\\slash", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := checkForPathTraversal(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "path traversal attempt detected")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_ListServices(t *testing.T) {
	tm := NewManager(nil)

	// Add Service Info
	service1 := &ServiceInfo{Name: "service-1", Config: &configv1.UpstreamServiceConfig{}}
	service2 := &ServiceInfo{Name: "service-2", Config: &configv1.UpstreamServiceConfig{}}

	tm.AddServiceInfo("id-1", service1)
	tm.AddServiceInfo("id-2", service2)

	services := tm.ListServices()
	assert.Len(t, services, 2)

	// Check content
	names := make(map[string]bool)
	for _, s := range services {
		names[s.Name] = true
	}
	assert.True(t, names["service-1"])
	assert.True(t, names["service-2"])
}

func TestCommandTool_Execute_PathTraversal_Args(t *testing.T) {
	// Setup command tool with args injection vulnerability
	toolProto := &v1.Tool{
		Name: proto.String("cmd-tool"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"{{arg}}"},
	}

	cmdTool := NewCommandTool(toolProto, service, callDef, nil, "")

	// Test path traversal in args
	req := &ExecutionRequest{
		ToolName: "cmd-tool",
		ToolInputs: []byte(`{"arg": "../etc/passwd"}`),
	}

	_, err := cmdTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestCommandTool_Execute_PathTraversal_Env(t *testing.T) {
	// Setup command tool with env injection vulnerability
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"env_var": map[string]interface{}{},
		},
	})

	toolProto := &v1.Tool{
		Name: proto.String("cmd-tool"),
		Annotations: &v1.ToolAnnotations{
			InputSchema: inputSchema,
		},
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
	}

	// Parameter mapping to env
	schema := &configv1.ParameterSchema{Name: proto.String("env_var")}
	mapping := &configv1.CommandLineParameterMapping{
		Schema: schema,
	}

	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{mapping},
	}

	cmdTool := NewCommandTool(toolProto, service, callDef, nil, "")

	// Test path traversal in env var (which checks validation)
	req := &ExecutionRequest{
		ToolName: "cmd-tool",
		ToolInputs: []byte(`{"env_var": "../bad"}`),
	}

	_, err := cmdTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}
