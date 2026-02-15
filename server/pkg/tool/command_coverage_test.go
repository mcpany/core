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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Timeout: durationpb.New(2 * time.Second),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-n"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

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
		pb.Tool_builder{Name: proto.String("echo-tool"), InputSchema: inputSchema}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("cat"), // cat echoes stdin to stdout
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Timeout:               durationpb.New(2 * time.Second),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewLocalCommandTool(
		pb.Tool_builder{Name: proto.String("json-tool")}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("false"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewLocalCommandTool(
		pb.Tool_builder{Name: proto.String("fail-tool")}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("nonexistentcommand_xyz"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewLocalCommandTool(
		pb.Tool_builder{Name: proto.String("bad-tool")}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sleep"),
		Timeout: durationpb.New(100 * time.Millisecond),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"1"}, // Sleep 1 second
	}.Build()

	tool := NewLocalCommandTool(
		pb.Tool_builder{Name: proto.String("timeout-tool")}.Build(),
		svc,
		callDef,
		nil,
		"call-id",
	)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	res, err := tool.Execute(context.Background(), req)

	if err != nil {
		// Can happen if timeout occurs during command startup
		assert.Contains(t, err.Error(), "context deadline exceeded")
	} else {
		resMap, ok := res.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, -1, resMap["return_code"])
	}
}

func TestLocalCommandTool_Execute_InvalidArgs(t *testing.T) {
	t.Parallel()
	svc := configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}.Build()
	tool := NewLocalCommandTool(pb.Tool_builder{Name: proto.String("t")}.Build(), svc, configv1.CommandLineCallDefinition_builder{}.Build(), nil, "call-id")

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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Timeout: durationpb.New(2 * time.Second),
		// ContainerEnvironment nil -> uses LocalExecutor
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-n"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

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
		pb.Tool_builder{Name: proto.String("echo-tool"), InputSchema: inputSchema}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("cat"),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Timeout:               durationpb.New(2 * time.Second),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewCommandTool(
		pb.Tool_builder{Name: proto.String("json-command-tool")}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("false"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewCommandTool(
		pb.Tool_builder{Name: proto.String("fail-command-tool")}.Build(),
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
	svc := configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}.Build()
	tool := NewCommandTool(pb.Tool_builder{Name: proto.String("t")}.Build(), svc, configv1.CommandLineCallDefinition_builder{}.Build(), nil, "call-id")

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
