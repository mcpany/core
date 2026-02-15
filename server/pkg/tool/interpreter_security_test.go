// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRubyCommandInjection_CodeFlag(t *testing.T) {
	// Define a tool that uses 'ruby', which is an interpreter.
	// We allow 'code' input.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("ruby-tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `ruby -e {{code}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: puts 12345
	// This contains spaces but no dangerous chars.
	payload := "puts 12345"

	req := &ExecutionRequest{
		ToolName: "ruby-tool",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	_, err := localTool.Execute(context.Background(), req)

    // Should fail because -e precedes {{code}} and space is considered dangerous for code arguments
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "injection detected")
}

func TestPythonCommandInjection_CodeFlag(t *testing.T) {
    // python -c {{code}}
    inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("python-tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: print 1
	payload := "print 1"

	req := &ExecutionRequest{
		ToolName: "python-tool",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	_, err := localTool.Execute(context.Background(), req)

    // Should fail because -c precedes {{code}}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "injection detected")
}

func TestPythonCommandInjection_NoCodeFlag(t *testing.T) {
    // python script.py {{arg}}
    inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"arg": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("python-script-tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"script.py", "{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: Hello World
	payload := "Hello World"

	req := &ExecutionRequest{
		ToolName: "python-script-tool",
		ToolInputs: []byte(`{"arg": "` + payload + `"}`),
	}

	_, err := localTool.Execute(context.Background(), req)

    // Should PASS because "script.py" is not a code flag, so spaces are allowed for data
	assert.NoError(t, err)
}
