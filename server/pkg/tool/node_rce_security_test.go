package tool

import (
	"context"
	"fmt"
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
	// Double quoted argument passed to node -e
	// node -e "console.log('{{code}}')"
	callDef.SetArgs([]string{"-e", "\"console.log('{{code}}')\""})
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

	// Attack payload: '); Function('return process')().exit(1)//
	// This closes the single quote (inside JS), executes Function, and comments out the rest.
	// But since it is inside double quotes (Shell), we can use single quotes freely.
	payload := "'); Function('return process')().exit(1)//"

	jsonInput := fmt.Sprintf(`{"code": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "node_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

    t.Logf("Error: %v", err)

	if err == nil {
        assert.Fail(t, "Function injection was not blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
