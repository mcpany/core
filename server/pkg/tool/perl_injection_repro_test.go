package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PerlInjection(t *testing.T) {
	// Define a tool that runs perl -e 'open(F, "{{msg}}"); print <F>'
	tool := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Template uses double quotes for the filename argument.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "open(F, \"{{msg}}\"); print <F>"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: | echo INJECTED
	// Result: open(F, "| echo INJECTED");
	// This executes the command and reads output.
	payload := "| echo INJECTED"

	inputs := map[string]interface{}{
		"msg": payload,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "perl-tool",
		ToolInputs: json.RawMessage(inputBytes),
	}

	result, err := localTool.Execute(context.Background(), req)

	// Expect the execution to be blocked by security checks
	if err == nil {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, _ := resMap["stdout"].(string)
		t.Logf("Stdout: %s", stdout)
		t.Fatal("Injection WAS NOT BLOCKED! Vulnerability persists.")
	}

	t.Logf("Blocked as expected. Error: %v", err)
	if !strings.Contains(err.Error(), "interpreter injection detected") {
		t.Fatalf("Blocked but with unexpected error message: %v", err)
	}
}
