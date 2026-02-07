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

func TestPhpRCE(t *testing.T) {
	cmd := "php"

	// Tool: php -r 'echo "{{msg}}";'
	// Double quoted string in PHP.

	toolDef := v1.Tool_builder{Name: proto.String("php-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "echo \"{{msg}}\";"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: { ${ system('echo RCE_SUCCESS') } }
	// We use spaces to bypass strings.Contains(val, "${") check.
	// But PHP allows spaces inside { ${ ... } }.

	injection := "{ ${ system('echo RCE_SUCCESS') } }"

	inputs := map[string]string{
		"msg": injection,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "php-tool",
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
			t.Errorf("FAIL: PHP RCE Succeeded!")
		}
	}
}
