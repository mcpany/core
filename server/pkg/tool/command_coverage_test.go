// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- LocalCommandTool Tests ---

func TestLocalCommandTool_Execute_Echo(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Timeout: durationpb.New(2 * time.Second),
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-n"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("args")}},
		},
	}

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	tool := NewLocalCommandTool(
		&pb.Tool{Name: proto.String("echo-tool"), InputSchema: inputSchema},
		svc,
		callDef,
		nil,
		"call-id",
	)

	// Execute adds input args
	req := &ExecutionRequest{
		ToolName:   "echo-tool",
		ToolInputs: json.RawMessage(`{"args": ["hello"]}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello", resMap["stdout"])
	assert.Equal(t, 0, resMap["return_code"])
}

func TestLocalCommandTool_Execute_JSON(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command:               proto.String("cat"), // cat echoes stdin to stdout
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Timeout:               durationpb.New(2 * time.Second),
	}
	callDef := &configv1.CommandLineCallDefinition{}

	tool := NewLocalCommandTool(
		&pb.Tool{Name: proto.String("json-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	// Input JSON will be echoed back.
	// cliExecutor expects output to be JSON.
	req := &ExecutionRequest{
		ToolName:   "json-tool",
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", resMap["key"])
}

func TestLocalCommandTool_Execute_Error(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("false"),
	}
	callDef := &configv1.CommandLineCallDefinition{}

	tool := NewLocalCommandTool(
		&pb.Tool{Name: proto.String("fail-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err) // It returns result with non-zero exit code

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEqual(t, 0, resMap["return_code"])
}

func TestLocalCommandTool_Execute_NonExistent(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("nonexistentcommand_xyz"),
	}
	callDef := &configv1.CommandLineCallDefinition{}

	tool := NewLocalCommandTool(
		&pb.Tool{Name: proto.String("bad-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executable file not found")
}

func TestLocalCommandTool_Execute_Timeout(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("sleep"),
		Timeout: durationpb.New(100 * time.Millisecond),
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"1"}, // Sleep 1 second
	}

	tool := NewLocalCommandTool(
		&pb.Tool{Name: proto.String("timeout-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, -1, resMap["return_code"])
}

func TestLocalCommandTool_Execute_InvalidArgs(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{Command: proto.String("echo")}
	tool := NewLocalCommandTool(&pb.Tool{Name: proto.String("t")}, svc, &configv1.CommandLineCallDefinition{}, nil, "call-id")

	// args is not array
	_, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"args": "not-array"}`)})
	assert.Error(t, err)

	// args element not string
	_, err = tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"args": [123]}`)})
	assert.Error(t, err)
}

// --- CommandTool Tests (using Local Executor implicitly) ---

func TestCommandTool_Execute_Echo(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Timeout: durationpb.New(2 * time.Second),
		// ContainerEnvironment nil -> uses LocalExecutor
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-n"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("args")}},
		},
	}

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	tool := NewCommandTool(
		&pb.Tool{Name: proto.String("echo-tool"), InputSchema: inputSchema},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{
		ToolName:   "echo-tool",
		ToolInputs: json.RawMessage(`{"args": ["hello-command-tool"]}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello-command-tool", resMap["stdout"])
}

func TestCommandTool_Execute_JSON(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command:               proto.String("cat"),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Timeout:               durationpb.New(2 * time.Second),
	}
	callDef := &configv1.CommandLineCallDefinition{}

	tool := NewCommandTool(
		&pb.Tool{Name: proto.String("json-command-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{
		ToolName:   "json-command-tool",
		ToolInputs: json.RawMessage(`{"cmd": "tool"}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "tool", resMap["cmd"])
}

func TestCommandTool_Execute_Error(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{
		Command: proto.String("false"),
	}
	callDef := &configv1.CommandLineCallDefinition{}

	tool := NewCommandTool(
		&pb.Tool{Name: proto.String("fail-command-tool")},
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEqual(t, 0, resMap["return_code"])
}

func TestCommandTool_Execute_InvalidArgs(t *testing.T) {
	t.Parallel()
	svc := &configv1.CommandLineUpstreamService{Command: proto.String("echo")}
	tool := NewCommandTool(&pb.Tool{Name: proto.String("t")}, svc, &configv1.CommandLineCallDefinition{}, nil, "call-id")

	// args is not array
	_, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"args": "not-array"}`)})
	assert.Error(t, err)

	// args element not string
	_, err = tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"args": [123]}`)})
	assert.Error(t, err)
}

func TestPrettyPrint(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", prettyPrint(nil, ""))
	assert.Contains(t, prettyPrint([]byte{0, 1, 2}, "application/octet-stream"), "Binary Data")
	assert.Contains(t, prettyPrint([]byte(`{"a":1}`), "application/json"), `"a": 1`)
	assert.Equal(t, `{"a":`, prettyPrint([]byte(`{"a":`), "application/json")) // Invalid JSON returns raw

	xmlData := `<root><a>1</a></root>`
	assert.Contains(t, prettyPrint([]byte(xmlData), "application/xml"), `<a>1</a>`)
	assert.Equal(t, `<invalid>`, prettyPrint([]byte(`<invalid>`), "application/xml"))
}
