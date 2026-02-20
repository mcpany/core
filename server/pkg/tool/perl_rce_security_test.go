package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPerlRCE_Unquoted(t *testing.T) {
	// Define a tool that runs perl -e {{script}}
	toolDef := v1.Tool_builder{
		Name: proto.String("perl_runner"),
	}.Build()

	cmd := "perl"
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload that avoids forbidden characters: () " ' ; etc.
	// system ls
	// in perl, system ls is system('ls') because ls is a bareword.

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "system bareword",
			payload: "system ls",
		},
		{
			name:    "open pipe",
			payload: "open F, |ls",
		},
		{
			name:    "exec bareword",
			payload: "exec ls",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "perl_runner",
				ToolInputs: []byte(fmt.Sprintf(`{"script": %q}`, tc.payload)),
			}

			_, err := tool.Execute(context.Background(), req)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "injection detected", "Payload %q should be blocked", tc.payload)
		})
	}
}
