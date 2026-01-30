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

func TestInterpreterInjection(t *testing.T) {
	// 1. Ruby Interpolation Injection
	t.Run("Ruby_DoubleQuotes_Injection", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name: proto.String("ruby_tool"),
		}
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("ruby"),
		}).Build()

		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-e", "puts \"{{msg}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{
						Name: proto.String("msg"),
					},
				},
			},
		}

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

		// Expect error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ruby interpolation injection detected")
	})

	// 2. Python f-string Injection
	t.Run("Python_FString_Injection", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name: proto.String("python_tool"),
		}
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("python3"),
		}).Build()

		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-c", "print(f'{{msg}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{
						Name: proto.String("msg"),
					},
				},
			},
		}

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
		assert.Contains(t, err.Error(), "python f-string injection detected")
	})

	// 3. Python fr-string Injection (Raw F-String)
	t.Run("Python_RawFString_Injection", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name: proto.String("python_tool"),
		}
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("python3"),
		}).Build()

		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-c", "print(fr'{{msg}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{
						Name: proto.String("msg"),
					},
				},
			},
		}

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
		assert.Contains(t, err.Error(), "python f-string injection detected")
	})

    // 4. Valid JSON input (Should be allowed)
    t.Run("Python_Valid_JSON", func(t *testing.T) {
        toolDef := &pb.Tool{
            Name: proto.String("python_json"),
        }
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("python3"),
		}).Build()

        callDef := &configv1.CommandLineCallDefinition{
            Args: []string{"process.py", "'{{msg}}'"},
            Parameters: []*configv1.CommandLineParameterMapping{
                {
                    Schema: &configv1.ParameterSchema{
                        Name: proto.String("msg"),
                    },
                },
            },
        }

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

        if err != nil {
             assert.NotContains(t, err.Error(), "injection detected", "Valid JSON should not be flagged as injection")
        }
    })

    // 5. PHP Valid Email (Should be allowed)
    t.Run("PHP_Valid_Email", func(t *testing.T) {
        toolDef := &pb.Tool{
            Name: proto.String("php_tool"),
        }
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("php"),
		}).Build()

        callDef := &configv1.CommandLineCallDefinition{
            Args: []string{"-r", "echo \"{{email}}\""},
            Parameters: []*configv1.CommandLineParameterMapping{
                {
                    Schema: &configv1.ParameterSchema{
                        Name: proto.String("email"),
                    },
                },
            },
        }

        tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

        input := `test@example.com` // Contains @

        req := &ExecutionRequest{
            ToolName: "php_tool",
            ToolInputs: []byte(fmt.Sprintf(`{"email": %q}`, input)),
            Arguments: map[string]interface{}{
                "email": input,
            },
        }

        _, err := tool.Execute(context.Background(), req)

        if err != nil {
             assert.NotContains(t, err.Error(), "injection detected", "Valid email should not be flagged as injection in PHP")
        }
    })
}
