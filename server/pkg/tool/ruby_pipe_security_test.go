package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRubyOpenInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where Ruby open() is used.

	// We simulate a tool that opens a file/URL provided by user.
	// E.g. ruby -e 'open("{{url}}").read'

	service := &configv1.CommandLineUpstreamService{}
	service.SetCommand("ruby")

	callDef := &configv1.CommandLineCallDefinition{}
	// Note: open() is deprecated in Ruby 3.x for Kernel#open, but still exists.
	// URI.open is safer but open("|cmd") might still work if not restricted.
	// We use -e 'puts open("{{url}}").read'
	callDef.SetArgs([]string{"-e", "require 'open-uri'; puts open(\"{{url}}\").read"})

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("url")

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	toolProto := &pb.Tool{}
	toolProto.SetName("ruby_read")

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload 1: |echo RCE_SUCCESS (Should be BLOCKED)
	payload1 := "|echo RCE_SUCCESS"
	req1 := &ExecutionRequest{
		ToolName:   "ruby_read",
		ToolInputs: []byte(`{"url": "` + payload1 + `"}`),
	}

	_, err1 := tool.Execute(context.Background(), req1)
	if err1 != nil {
		t.Logf("Execution blocked as expected for '|...': %v", err1)
		assert.Contains(t, err1.Error(), "ruby injection detected")
	} else {
		t.Fatal("FAILED: Expected execution to be blocked for payload starting with '|'")
	}

	// Payload 2:  |echo RCE_SUCCESS (Leading space - Should be ALLOWED by security check, and Safe in Ruby)
	payload2 := " |echo RCE_SUCCESS"
	req2 := &ExecutionRequest{
		ToolName:   "ruby_read",
		ToolInputs: []byte(`{"url": "` + payload2 + `"}`),
	}

	result2, err2 := tool.Execute(context.Background(), req2)
	if err2 != nil {
		// It might fail with "No such file or directory" which is expected and SAFE.
		// It should NOT be blocked by "injection detected".
		t.Logf("Execution failed (runtime error, expected): %v", err2)
		assert.NotContains(t, err2.Error(), "ruby injection detected")
	} else {
		// If it succeeded (err2 == nil), it means exit code was 0.
		// This is unexpected for a failed open(), but if it happens, we ensure RCE didn't happen.
		resMap, ok := result2.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)
		t.Logf("Stdout2: %s", stdout)

		// It should NOT contain RCE output
		assert.NotContains(t, stdout, "RCE_SUCCESS", "Safe input ' |echo' executed RCE!")
	}
}
