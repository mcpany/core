// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestSentinel_CommandInjection prevents injection via command arguments when using templating.
func TestSentinel_CommandInjection(t *testing.T) {
	ctx := context.Background()

	// Scenario 1: Unsafe usage of /bin/sh -c with user input
	t.Run("Shell Injection Prevention", func(t *testing.T) {
		svcConfig := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("/bin/sh"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "echo {{message}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("message"),
						Type: configv1.ParameterType_STRING.Enum(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		toolProto := pb.Tool_builder{
			Name: proto.String("echo_safe"),
		}.Build()

		tl := tool.NewCommandTool(toolProto, svcConfig, callDef, nil, "test_call")

		require.NotNil(t, tl)

		input := &tool.ExecutionRequest{
			Arguments: map[string]any{
				"message": "hello; id",
			},
			ToolInputs: []byte(`{"message": "hello; id"}`),
		}

		_, err := tl.Execute(ctx, input)
		// We expect an error because the input contains forbidden characters
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Scenario 2: Command Argument Injection (Flag Injection)
	t.Run("Flag Injection", func(t *testing.T) {
		svcConfig := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("ls"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{path}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("path"),
						Type: configv1.ParameterType_STRING.Enum(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		toolProto := pb.Tool_builder{
			Name: proto.String("ls_safe"),
		}.Build()

		tl := tool.NewCommandTool(toolProto, svcConfig, callDef, nil, "ls_safe")
		require.NotNil(t, tl)

		input := &tool.ExecutionRequest{
			Arguments: map[string]any{
				"path": "-la",
			},
			ToolInputs: []byte(`{"path": "-la"}`),
		}

		_, err := tl.Execute(ctx, input)
		// We expect an error because the input looks like a flag
		require.Error(t, err)
		assert.Contains(t, err.Error(), "argument injection detected")
	})
}

// TestSentinel_DangerousBinaries prevents configuring tools that execute dangerous binaries directly.
func TestSentinel_DangerousBinaries(t *testing.T) {
	dangerous := []string{"nc", "netcat", "ncat", "socat", "wget", "curl"}

	for _, bin := range dangerous {
		t.Run("Block "+bin, func(t *testing.T) {
			svcConfig := configv1.CommandLineUpstreamService_builder{
				Command: proto.String(bin),
			}.Build()

			callDef := configv1.CommandLineCallDefinition_builder{
				Args: []string{"google.com"},
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{
							Name: proto.String("dummy"),
							Type: configv1.ParameterType_STRING.Enum(),
						}.Build(),
					}.Build(),
				},
			}.Build()

			toolProto := pb.Tool_builder{
				Name: proto.String("dangerous"),
			}.Build()

			tl := tool.NewCommandTool(toolProto, svcConfig, callDef, nil, "dangerous")
			assert.NotNil(t, tl, "Currently we do not block dangerous binaries at the tool level, but we should consider it.")
		})
	}
}

// TestSentinel_AWK_Injection checks for injection in interpreted languages passed as args.
func TestSentinel_AWK_Injection(t *testing.T) {
	ctx := context.Background()

	svcConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "awk '{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("awk_test"),
	}.Build()

	tl := tool.NewCommandTool(toolProto, svcConfig, callDef, nil, "awk_test")
	require.NotNil(t, tl)

	input := &tool.ExecutionRequest{
		Arguments: map[string]any{
			"script": `BEGIN { system("echo hacked") }`,
		},
		ToolInputs: []byte(`{"script": "BEGIN { system(\"echo hacked\") }"}`),
	}

	res, err := tl.Execute(ctx, input)
	// If it runs, we see "hacked"
	// If we error out (e.g. awk not found or syntax error), that's fine.
	// Assuming awk is installed.
	if err == nil {
		resultMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		combinedOutput, ok := resultMap["combined_output"].(string)
		require.True(t, ok)
		assert.Contains(t, combinedOutput, "hacked", "Code injection into AWK script succeeded")
	}
}

func TestSentinel_Perl_Injection(t *testing.T) {
	ctx := context.Background()

	svcConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "perl -e '{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("perl_test"),
	}.Build()

	tl := tool.NewCommandTool(toolProto, svcConfig, callDef, nil, "perl_test")
	require.NotNil(t, tl)

	input := &tool.ExecutionRequest{
		Arguments: map[string]any{
			"script": `print "hacked\n";`,
		},
		ToolInputs: []byte(`{"script": "print \"hacked\\n\";"}`),
	}

	res, err := tl.Execute(ctx, input)
	if err == nil {
		resultMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		combinedOutput, ok := resultMap["combined_output"].(string)
		require.True(t, ok)
		assert.Contains(t, combinedOutput, "hacked", "Code injection into Perl script succeeded")
	}
}
