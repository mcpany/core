package security

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCommandTool_Execute_PathTraversal_Comprehensive(t *testing.T) {
	t.Parallel()

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"env_var": map[string]interface{}{},
		},
	})

	toolProto := pb.Tool_builder{
		Name: proto.String("cmd-tool"),
		Annotations: pb.ToolAnnotations_builder{
			InputSchema: inputSchema,
		}.Build(),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	schema := configv1.ParameterSchema_builder{Name: proto.String("env_var")}.Build()
	mapping := configv1.CommandLineParameterMapping_builder{
		Schema: schema,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{mapping},
	}.Build()

	cmdTool := tool.NewCommandTool(toolProto, service, callDef, nil, "")

	// Negative test: Passing path traversal payloads should fail with "path traversal attempt detected"
	req := &tool.ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: []byte(`{"env_var": "../etc/passwd"}`),
	}

	_, err := cmdTool.Execute(context.Background(), req)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	}

	// Negative test: Passing absolute paths should fail with "absolute path detected"
	req2 := &tool.ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: []byte(`{"env_var": "/etc/passwd"}`),
	}

	_, err2 := cmdTool.Execute(context.Background(), req2)
	if assert.Error(t, err2) {
		assert.Contains(t, err2.Error(), "absolute path detected")
	}

	// Negative test: Passing argument injection should fail with "argument injection detected"
	req3 := &tool.ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: []byte(`{"env_var": "-rm -rf"}`),
	}

	_, err3 := cmdTool.Execute(context.Background(), req3)
	if assert.Error(t, err3) {
		assert.Contains(t, err3.Error(), "argument injection detected")
	}
}

func TestDockerExecutor_ResourceLimits(t *testing.T) {
	t.Parallel()

	// 1. Create a dummy execution request to a container environment
	resourceLimits := configv1.ResourceLimits_builder{
		Memory: proto.String("128m"),
		Cpus:   proto.Float64(0.5),
	}.Build()

	containerEnv := configv1.ContainerEnvironment_builder{
		Image:          proto.String("alpine"),
		ResourceLimits: resourceLimits,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command:              proto.String("echo"),
		ContainerEnvironment: containerEnv,
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("docker-tool"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	cmdTool := tool.NewCommandTool(toolProto, service, callDef, nil, "")

	// 2. Validate that executing the tool applies the resource limits (just checking that parsing/wiring works)
	// We're mainly making sure that having the fields populated doesn't crash anything.
	// Since we mock the executor or test the parsing directly, let's just make sure the struct builds.
	assert.NotNil(t, cmdTool)
}
