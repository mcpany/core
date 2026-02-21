package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPythonInjectionBypass(t *testing.T) {
    // Scenario: Shell wrapper invoking python with single-quoted argument
	t.Run("Shell_Python_SingleQuote_Bypass", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("shell_python_tool"),
		}).Build()
		cmd := "bash"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// bash -c "python3 -c '{{code}}'"
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "python3 -c '{{code}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Bypass attempt:
        // 1. Use double quotes instead of single quotes (allowed in single-quoted context)
        // 2. Use concatenation to hide "eval"
		input := `print(__builtins__["ev"+"al"]("1+1"))`

		req := &ExecutionRequest{
			ToolName: "shell_python_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err, "Should detect bypass via concatenation in shell single-quoted context")
	})

    // Scenario: Python tool directly (double quotes allowed if quoteLevel allows it, but keywords blocked)
    // Wait, checkUnquotedInjection blocks quotes for python.
    // So double quotes might be blocked if unquoted.
    // But checkUnquotedInjection is not called if quoteLevel is 1.
    // If args: ["-c", "\"{{code}}\""]. quoteLevel 1.
    t.Run("Python_DoubleQuote_Bypass", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("python_tool"),
		}).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

        // Double quoted argument
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "\"{{code}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Bypass attempt:
        // Use concat to hide eval.
        // Use unquoted __builtins__ (allowed in double quoted string in python if it's evaluated)
		input := `__builtins__["ev"+"al"]("print(1)")`

		req := &ExecutionRequest{
			ToolName: "python_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err, "Should detect bypass via concatenation in python double-quoted context")
	})
}
