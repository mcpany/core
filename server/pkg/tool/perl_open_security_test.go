package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Perl_Open_Injection(t *testing.T) {
	// Vulnerability: Perl allows 'open' without parentheses, e.g. open FH, ">file"
	// This might bypass checkInterpreterFunctionCalls which looks for open(

	toolDef := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
		Args: []string{"-e", "'{{code}}'"},
	}.Build()

	// NewLocalCommandTool(tool *v1.Tool, service *configv1.CommandLineUpstreamService, callDefinition *configv1.CommandLineCallDefinition, policies []*configv1.CallPolicy, callID string) Tool
	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Attack payload: open F, ">hacked";
	// This writes to a file named "hacked".
	// We also try "system 'ls'" just to verify protection works for known cases.

	testCases := []struct {
		name    string
		payload string
		blocked bool
	}{
		{
			name:    "Blocked system call",
			payload: `system "ls"`,
			blocked: true,
		},
		{
			name:    "Blocked open with parens",
			payload: `open(F, ">hacked")`,
			blocked: true,
		},
		{
			name:    "Vulnerable open without parens",
			payload: `open F, ">hacked"`,
			blocked: true, // Should be blocked, but we suspect it bypasses
		},
		{
			name:    "Vulnerable open pipe (RCE)",
			payload: `open F, "|ls"`,
			blocked: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputMap := map[string]interface{}{
				"code": tc.payload,
			}
			inputBytes, _ := json.Marshal(inputMap)

			req := &ExecutionRequest{
				ToolName: "perl-tool",
				ToolInputs: inputBytes,
			}

			_, err := localTool.Execute(context.Background(), req)

			if tc.blocked {
				if err == nil {
					t.Errorf("Payload %q should be blocked but was allowed", tc.payload)
				} else {
					assert.Contains(t, err.Error(), "injection detected", "Payload %q should be flagged as injection", tc.payload)
				}
			} else {
				// If expected allowed (not the case here, all are dangerous)
			}
		})
	}
}
