package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestInterpreterSecurityFixes(t *testing.T) {
	t.Run("PythonCodeInjection", func(t *testing.T) {
		cmd := "python3"
		toolDef := v1.Tool_builder{Name: proto.String("python-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "print(\"Hello {{name}}\")"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("name")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
		injection := "\"); import os; s=os.system; s('echo RCE_SUCCESS'); print(\""

		inputs := map[string]string{"name": injection}
		inputBytes, _ := json.Marshal(inputs)
		req := &ExecutionRequest{ToolName: "python-tool", ToolInputs: inputBytes}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "interpreter injection detected") && !strings.Contains(err.Error(), "shell injection detected") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("PythonDoubleQuoteRCE", func(t *testing.T) {
		cmd := "python3"
		toolDef := v1.Tool_builder{Name: proto.String("python-dq-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "print(\"Hello {{name}}\")"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("name")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
		// This payload previously bypassed the check because quoteLevel 1 didn't check dangerousCalls
		injection := "\"+__import__('os').system('echo RCE_CONFIRMED')+\""

		inputs := map[string]string{"name": injection}
		inputBytes, _ := json.Marshal(inputs)
		req := &ExecutionRequest{ToolName: "python-dq-tool", ToolInputs: inputBytes}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "interpreter injection detected") && !strings.Contains(err.Error(), "shell injection detected") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("PythonEvalRCE", func(t *testing.T) {
		cmd := "python3"
		toolDef := v1.Tool_builder{Name: proto.String("python-eval-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "eval(\"{{code}}\")"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
		injection := "__import__('os').system('echo RCE_EVAL_SUCCESS')"

		inputs := map[string]string{"code": injection}
		inputBytes, _ := json.Marshal(inputs)
		req := &ExecutionRequest{ToolName: "python-eval-tool", ToolInputs: inputBytes}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "interpreter injection detected") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("AwkFileRead", func(t *testing.T) {
		cmd := "awk"
		toolDef := v1.Tool_builder{Name: proto.String("awk-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{script}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
		injection := "BEGIN { while((getline < \"/etc/hosts\") > 0) print }"

		inputs := map[string]string{"script": injection}
		inputBytes, _ := json.Marshal(inputs)
		req := &ExecutionRequest{ToolName: "awk-tool", ToolInputs: inputBytes}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "awk injection detected") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}
