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

func TestFalsePositives(t *testing.T) {
	cmd := "python3"
	toolDef := v1.Tool_builder{Name: proto.String("fp-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print(\"{{input}}\")"}, // Level 1 Double Quoted
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	safeInputs := []string{
		"photos.jpg",
		"system analysis",
		"file system",
		"open door",
		"The quick brown fox",
		"sys.log",
	}

	for _, input := range safeInputs {
		inputMap := map[string]string{"input": input}
		inputBytes, _ := json.Marshal(inputMap)
		req := &ExecutionRequest{ToolName: "fp-tool", ToolInputs: inputBytes}

		res, err := tool.Execute(context.Background(), req)
		// We expect success (or at least NOT an injection error)
		// Since we run dry, it might execute python and print.
		// But we check err.
		if err != nil {
			if strings.Contains(err.Error(), "injection detected") {
				t.Errorf("False positive detected for input %q: %v", input, err)
			}
		} else {
			// Check output just in case
			resMap, _ := res.(map[string]interface{})
			combined, _ := resMap["combined_output"].(string)
			// stdout should contain input
			if !strings.Contains(combined, input) {
				// Maybe stderr?
			}
		}
	}
}
