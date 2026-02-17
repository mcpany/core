package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestPerlRCEInjection(t *testing.T) {
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("perl")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	// Unquoted injection!
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

	payloads := []string{
		"readpipe qw/ls/",
		"syscall 1",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			jsonInput := fmt.Sprintf(`{"name": "%s"}`, payload)

			req := &ExecutionRequest{
				ToolName:   "perl_rce",
				ToolInputs: []byte(jsonInput),
			}

			_, err := tool.Execute(ctx, req)

			if err != nil {
				t.Logf("Error: %s", err.Error())
			}

			if err == nil {
				// Vulnerability confirmed
				assert.Fail(t, fmt.Sprintf("Injection '%s' was not blocked", payload))
			} else {
				assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
			}
		})
	}
}
