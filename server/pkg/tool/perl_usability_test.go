package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PerlUsability(t *testing.T) {
	// Define a tool that uses 'perl'
	tool := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `perl {{arg}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	tests := []struct {
		name      string
		arg       string
		shouldPass bool
	}{
		{"system.pl", "system.pl", true},
		{"system-config", "system-config", true},
		{"system/path", "system/path", true},
		{"system+id", "system+id", false}, // + is blocked because it allows RCE (system+q/id/)
		{"system=val", "system=val", true},
		{"system space", "system id", false},
		{"system parens", "system(id)", false},
		{"system brace", "system{id}", false},
		{"system bracket", "system[id]", false},
		{"system standalone", "system", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "perl-tool",
				Arguments: map[string]interface{}{
					"arg": tt.arg,
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			_, err := localTool.Execute(context.Background(), req)

			if tt.shouldPass {
				if err != nil {
					t.Errorf("Expected argument %q to be allowed, but got error: %v", tt.arg, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected argument %q to be blocked, but it was allowed", tt.arg)
				} else {
					// Optional: check error message
					t.Logf("Correctly blocked %q: %v", tt.arg, err)
				}
			}
		})
	}
}
