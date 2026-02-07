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

func TestPerlRCE(t *testing.T) {
	cmd := "perl"

	// Tool: perl -e "print \"{{msg}}\""
	// Double quoted string in Perl.

	toolDef := v1.Tool_builder{Name: proto.String("perl-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print \"{{msg}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: @{[system('echo RCE_SUCCESS')]}
	// This uses Perl's array interpolation to execute code.
	// We use single quotes inside because double quotes are blocked by checkForShellInjection QuoteLevel 1.
	// @, {, [, ], (, ), ' are ALL ALLOWED in QuoteLevel 1.

	injection := "@{[system('echo RCE_SUCCESS')]}"

	inputs := map[string]string{
		"msg": injection,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "perl-tool",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		t.Logf("Blocked: %v", err)
	} else {
		resMap, _ := result.(map[string]interface{})
		stdout, _ := resMap["stdout"].(string)
		t.Logf("Stdout: %s", stdout)
		if strings.Contains(stdout, "RCE_SUCCESS") {
			t.Errorf("FAIL: Perl RCE Succeeded!")
		}
	}
}
