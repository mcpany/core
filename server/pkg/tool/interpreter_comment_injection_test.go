package tool

import (
	"context"
	"encoding/json"
	"testing"
    "strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestInterpreterCommentInjection(t *testing.T) {
	// Setup a PHP tool vulnerable to injection via argument substitution
	tool := pb.Tool_builder{
		Name: proto.String("php_eval"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "eval(\"{{input}}\");"}, // Vulnerable context: double quotes
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "test-call")

	// Payload that attempts to bypass checkInterpreterFunctionCalls
	// It uses a comment '#' to break the adjacency of 'system' and '('
	// This payload is: system # comment \n ('ls')
	// The newline is critical for PHP comment syntax to end the comment.
	payload := "system # comment\n('ls')"

	inputs := map[string]interface{}{
		"input": payload,
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "php_eval",
		ToolInputs: inputsBytes,
		DryRun:     true,
	}

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("Vulnerability Reproduced: Validation bypassed! Expected error 'interpreter injection detected'")
	}

    if !strings.Contains(err.Error(), "interpreter injection detected") {
        t.Errorf("Expected 'interpreter injection detected' error, got: %v", err)
    }
}
