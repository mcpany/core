package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PerlRCE_Prevention(t *testing.T) {
	// Define a tool that uses 'perl'
	tool := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `perl -e {{script}}` (unquoted)
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	// We disable policies for this test to focus on the injection check
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: system q/id/
	// This uses 'q//' operator for quoting strings in Perl, avoiding quotes that might be caught.
	// Space is allowed because perl is not considered a shell.
	payload := "system q/id/"

	req := &ExecutionRequest{
		ToolName: "perl-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, the protection failed.
	if err == nil {
		t.Logf("Vulnerability Reproduced: Perl RCE payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}

func TestLocalCommandTool_PerlBypass_Plus(t *testing.T) {
	// Define a tool that uses 'perl'
	tool := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `perl -e {{script}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: system+("ls")
	// If + is whitelisted, this might pass checkKeyword.
	// But it contains ( and ), so checkUnquotedInjection should block it?
	payload := `system+("ls")`

	req := &ExecutionRequest{
		ToolName: "perl-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We expect this to be blocked by checkUnquotedInjection (due to parens) OR checkInterpreterFunctionCalls (due to system)
	if err == nil {
		t.Logf("Vulnerability Reproduced: Perl RCE payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}

func TestLocalCommandTool_PerlBypass_Plus_Space(t *testing.T) {
    // If we use system + "ls" (no parens)
    // " is blocked by checkUnquotedInjection.

    // What if we use q//?
    // system+q/ls/

	tool := v1.Tool_builder{Name: proto.String("perl-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{Command: proto.String("perl"), Local: proto.Bool(true)}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

    payload := `system+q/ls/`
    req := &ExecutionRequest{
		ToolName: "perl-tool",
		Arguments: map[string]interface{}{"script": payload},
	}
    req.ToolInputs, _ = json.Marshal(req.Arguments)
    _, err := localTool.Execute(context.Background(), req)

    if err == nil {
        t.Logf("Vulnerability Reproduced: Perl RCE payload %q was allowed!", payload)
        t.Fail()
    } else {
        t.Logf("Blocked: %v", err)
    }
}
