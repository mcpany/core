package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Execute_PythonInjection_Blocked(t *testing.T) {
	// This test demonstrates that python IS treated as a shell command,
	// blocking code injection via argument substitution.

	toolDef := &v1.Tool{
		Name: proto.String("python_tool"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("python3"),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("msg"),
				},
			},
		},
	}

	ct := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-id")

	// Malicious input trying to break out of python string
	injectionPayload := "'); print(\"INJECTED\"); print('"
	jsonInput, _ := json.Marshal(map[string]string{"msg": injectionPayload})

	req := &ExecutionRequest{
		ToolName: "python_tool",
		ToolInputs: jsonInput,
	}

	// Execute
	_, err := ct.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error due to injection detection, but got nil")
	}

	if !strings.Contains(err.Error(), "shell injection detected") {
		t.Fatalf("Expected 'shell injection detected' error, got: %v", err)
	}
}

func TestLocalCommandTool_Execute_DockerInjection_Blocked(t *testing.T) {
	// This test demonstrates that docker IS treated as a shell command,
	// blocking injection.

	toolDef := &v1.Tool{
		Name: proto.String("docker_tool"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("docker"),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"run", "ubuntu", "bash", "-c", "echo {{msg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("msg"),
				},
			},
		},
	}

	ct := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-id")

	// Malicious input
	injectionPayload := "; rm -rf /"
	jsonInput, _ := json.Marshal(map[string]string{"msg": injectionPayload})

	req := &ExecutionRequest{
		ToolName: "docker_tool",
		ToolInputs: jsonInput,
	}

	// Execute
	_, err := ct.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error due to injection detection, but got nil")
	}

	if !strings.Contains(err.Error(), "shell injection detected") {
		t.Fatalf("Expected 'shell injection detected' error, got: %v", err)
	}
}
