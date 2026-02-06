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

func TestInterpreterSecurity(t *testing.T) {
	// 1. Ruby Interpolation Injection
	t.Run("Ruby_DoubleQuotes_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("ruby_tool"),
		}).Build()
		cmd := "ruby"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-e", "puts \"{{msg}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using Ruby interpolation
		input := "#{system('echo injected')}"

		req := &ExecutionRequest{
			ToolName: "ruby_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": "%s"}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ruby interpolation injection detected", "Should detect ruby interpolation")
	})

	// 2. Python F-String Injection
	t.Run("Python_FString_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("python_tool"),
		}).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "print(f'{{msg}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Input: {__import__("os").system("echo injected")}
		input := `{__import__("os").system("echo injected")}`

		req := &ExecutionRequest{
			ToolName: "python_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": %q}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "python f-string injection detected", "Should detect python f-string injection")
	})

    // 2b. Python Raw F-String Injection (fr'...')
    t.Run("Python_RawFString_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("python_tool"),
		}).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "print(fr'{{msg}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		input := `{__import__("os").system("echo injected")}`

		req := &ExecutionRequest{
			ToolName: "python_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": %q}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "python f-string injection detected", "Should detect python raw f-string injection")
	})

    // 3. Python Valid JSON input (Should be allowed)
    t.Run("Python_Valid_JSON", func(t *testing.T) {
        toolDef := (&pb.Tool_builder{
            Name: proto.String("python_json"),
        }).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

        callDef := (&configv1.CommandLineCallDefinition_builder{
            Args: []string{"process.py", "'{{msg}}'"},
            Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
            },
        }).Build()

        tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

        input := `{"foo": "bar"}` // Contains ", {, }

        req := &ExecutionRequest{
            ToolName: "python_json",
            ToolInputs: []byte(fmt.Sprintf(`{"msg": %q}`, input)),
            Arguments: map[string]interface{}{
                "msg": input,
            },
        }

        _, err := tool.Execute(context.Background(), req)

        // We expect checks to PASS. Execution might fail (process.py missing), but not "injection detected".
        if err != nil {
             assert.NotContains(t, err.Error(), "injection detected", "Valid JSON should not be flagged as injection")
        }
    })

	// 4. JS Template Literal Injection
	t.Run("JS_Template_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("node_tool"),
		}).Build()
		cmd := "node"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template uses backticks: console.log(`{{msg}}`)
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-e", "console.log(`{{msg}}`)"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		input := "${process.exit(1)}"

		req := &ExecutionRequest{
			ToolName: "node_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": %q}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "javascript template literal injection detected", "Should detect JS template injection")
	})

	// 5. Bash Backtick Injection
	t.Run("Bash_Backtick_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("bash_tool"),
		}).Build()
		cmd := "bash"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template uses backticks: echo `{{msg}}`
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "echo `{{msg}}`"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Attempt full shell injection using ;
		input := "; date"

		req := &ExecutionRequest{
			ToolName: "bash_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": %q}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected", "Should detect dangerous character in backticks for shell")
	})
}
