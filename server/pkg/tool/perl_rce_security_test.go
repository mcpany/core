package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestPerlReadpipeInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("perl")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	// Unquoted context test
	callDef.SetArgs([]string{"-e", "print {{name}}"})
	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolStruct := &v1.Tool{}
	toolStruct.SetName("perl_rce")

	tool := NewLocalCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	// Attack payload: readpipe ls
	// Resulting Perl code: print readpipe ls
	payload := "readpipe ls"

	jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "perl_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	// We expect an error containing "injection detected"
	assert.ErrorContains(t, err, "injection detected", "Expected injection detection error")
}
