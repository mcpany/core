package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_BashAwkFileWrite_Repro(t *testing.T) {
	// Define a tool that uses 'bash' to run 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("bash-awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `bash -c "awk '{{script}}'"`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "awk '{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attempt to write to a file in /tmp using awk redirection via bash
	tmpFile := "/tmp/bash_awk_pwned.txt"
	os.Remove(tmpFile) // Ensure clean state

	// Payload: BEGIN { print "pwned" > "/tmp/bash_awk_pwned.txt" }
	payload := `BEGIN { print "pwned" > "` + tmpFile + `" }`

	req := &ExecutionRequest{
		ToolName: "bash-awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	t.Logf("Executing with payload: %s", payload)
	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, check if file exists
	if err == nil {
		if _, err := os.Stat(tmpFile); err == nil {
			t.Logf("Vulnerability Reproduced: Awk successfully wrote to %s via bash", tmpFile)
			t.Fail()
			os.Remove(tmpFile)
		} else {
			// Maybe execution failed for other reasons (awk not found?)
			t.Logf("Execution succeeded but file not created? Check if awk/bash are available.")
		}
	} else {
		t.Logf("Blocked: %v", err)
	}
}
