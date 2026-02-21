package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSqlite3ShellInjection(t *testing.T) {
	// Setup sqlite3 tool
	toolProto := v1.Tool_builder{
		Name: proto.String("sqlite3"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sqlite3"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{":memory:", "{{query}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("query"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-call")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputs := map[string]interface{}{
		"query": ".shell echo pwned",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "sqlite3",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(ctx, req)

	// Now we expect an error!
	if err == nil {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		stdout, _ := resMap["stdout"].(string)
		if strings.Contains(stdout, "pwned") {
			t.Fatalf("VULNERABILITY CONFIRMED: sqlite3 .shell injection worked (Fix FAILED)")
		} else {
			t.Logf("Execution succeeded but payload didn't work. Check logic.")
		}
	} else {
		t.Logf("Execution failed as expected: %v", err)
		if !strings.Contains(err.Error(), "sqlite3 injection detected") {
			t.Fatalf("Unexpected error message: %v", err)
		}
	}
}

func TestSqlite3FalsePositive(t *testing.T) {
	// Setup sqlite3 tool
	toolProto := v1.Tool_builder{
		Name: proto.String("sqlite3"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sqlite3"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{":memory:", "{{query}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("query"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-call-safe")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a safe query where .shell is inside a string literal
	// We use VALUES instead of INSERT to avoid triggering the keyword blocker,
	// focusing the test on the .shell detection logic.
	inputs := map[string]interface{}{
		"query": "VALUES('User attempted .shell command');",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "sqlite3",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(ctx, req)

	// This might fail due to other security checks (like parens or keywords),
	// but it MUST NOT fail due to the .shell command check.
	if err != nil {
		if strings.Contains(err.Error(), "sqlite3 injection detected") {
			t.Fatalf("False positive detected! Safe query was blocked by sqlite3 check: %v", err)
		}
		t.Logf("Blocked by other checks (acceptable): %v", err)
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		_ = resMap
	}
}

func TestGdbNewlineInjection(t *testing.T) {
	// Setup gdb tool
	toolProto := v1.Tool_builder{
		Name: proto.String("gdb"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gdb"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-batch", "-ex", "'{{command}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("command"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, serviceConfig, callDef, nil, "test-gdb")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Injection payload: multiline escape
	inputs := map[string]interface{}{
		"command": "print 1\n!echo pwned",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "gdb",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(ctx, req)

	// Should be blocked
	if err == nil {
		t.Fatalf("GDB newline injection succeeded (should be blocked)")
	} else {
		if !strings.Contains(err.Error(), "gdb injection detected") {
			t.Logf("Blocked by other checks (acceptable): %v", err)
		}
	}
	_ = result
}
