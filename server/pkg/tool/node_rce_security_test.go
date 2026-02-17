package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestNodeFunctionInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("node")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("code")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	// Unquoted injection
	callDef.SetArgs([]string{"-e", "{{code}}"})
	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolStruct := &v1.Tool{}
	toolStruct.SetName("node_rce")

	tool := NewLocalCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	// Attack payload: new Function("return process")
    // This is unquoted injection.
	payload := `new Function("return process")`

	inputMap := map[string]string{"code": payload}
	jsonInput, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "node_rce",
		ToolInputs: jsonInput,
	}

	_, err := tool.Execute(ctx, req)

    if err != nil {
        t.Logf("Error: %s", err.Error())
    }

	if err == nil {
        assert.Fail(t, "Node Function injection was not blocked")
	} else {
        // Assert specific error message to verify my fix is active
		assert.Contains(t, err.Error(), "javascript Function constructor injection detected", "Expected Function constructor detection error")
	}
}
