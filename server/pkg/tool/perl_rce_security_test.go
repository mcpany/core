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
	// Unquoted argument! This allows passing arbitrary code to perl -e.
	// quoteLevel will be 0.
	callDef.SetArgs([]string{"-e", "{{name}}"})
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

	// Attack payload: readpipe qw/echo INJECTED/
	// This uses readpipe with qw// to avoid quotes and parentheses.
	// It only uses allowed characters for unquoted injection (no ; ( ) ' " etc).
	payload := "print readpipe qw/echo INJECTED/"

	jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "perl_rce",
		ToolInputs: []byte(jsonInput),
	}

	// We expect this to fail with "injection detected".
	_, err := tool.Execute(ctx, req)

    t.Logf("Error: %v", err)

	if err == nil {
		// Vulnerability confirmed: readpipe was NOT blocked!
        assert.Fail(t, "readpipe injection was not blocked")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
