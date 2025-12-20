// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/mcpany/core/pkg/command"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockExecutor is a mock for command.Executor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx context.Context, cmd string, args []string, dir string, env []string) (stdout, stderr io.ReadCloser, exitCode <-chan int, err error) {
	callArgs := m.Called(ctx, cmd, args, dir, env)
	return callArgs.Get(0).(io.ReadCloser), callArgs.Get(1).(io.ReadCloser), callArgs.Get(2).(<-chan int), callArgs.Error(3)
}

func (m *MockExecutor) ExecuteWithStdIO(ctx context.Context, cmd string, args []string, dir string, env []string) (stdin io.WriteCloser, stdout, stderr io.ReadCloser, exitCode <-chan int, err error) {
	callArgs := m.Called(ctx, cmd, args, dir, env)
	return callArgs.Get(0).(io.WriteCloser), callArgs.Get(1).(io.ReadCloser), callArgs.Get(2).(io.ReadCloser), callArgs.Get(3).(<-chan int), callArgs.Error(4)
}

type mockWriteCloser struct {
	io.Writer
}

func (m *mockWriteCloser) Close() error {
	return nil
}

func TestCommandTool_Execute_Success_Mock(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("cmd-tool")

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"hello", "{{name}}"},
	}

	mockExecutor := new(MockExecutor)

	// Setup mock return values
	mockStdout := io.NopCloser(bytes.NewBufferString("hello world\n"))
	mockStderr := io.NopCloser(bytes.NewBufferString(""))
	exitCodeChan := make(chan int, 1)
	exitCodeChan <- 0

	mockExecutor.On("Execute", mock.Anything, "echo", []string{"hello", "world"}, "", mock.Anything).Return(
		mockStdout, mockStderr, (<-chan int)(exitCodeChan), nil,
	)

	tool := &CommandTool{
		tool:           toolProto,
		service:        service,
		callDefinition: callDef,
		executorFactory: func(ce *configv1.ContainerEnvironment) command.Executor {
			return mockExecutor
		},
	}

	req := &ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: json.RawMessage(`{"name": "world"}`),
	}

	result, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello world\n", resMap["stdout"])
	assert.Equal(t, 0, resMap["return_code"])
}

func TestCommandTool_Execute_JSONProtocol_Mock(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("cmd-tool-json")

	service := &configv1.CommandLineUpstreamService{
		Command:               proto.String("processor"),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}

	callDef := &configv1.CommandLineCallDefinition{}

	mockExecutor := new(MockExecutor)

	// Setup mock
	mockStdin := &mockWriteCloser{Writer: bytes.NewBuffer(nil)}
	mockStdout := io.NopCloser(bytes.NewBufferString(`{"result": "success"}`))
	mockStderr := io.NopCloser(bytes.NewBufferString(""))
	exitCodeChan := make(chan int, 1)
	exitCodeChan <- 0

	mockExecutor.On("ExecuteWithStdIO", mock.Anything, "processor", []string{}, "", mock.Anything).Return(
		(io.WriteCloser)(mockStdin), mockStdout, mockStderr, (<-chan int)(exitCodeChan), nil,
	)

	tool := &CommandTool{
		tool:           toolProto,
		service:        service,
		callDefinition: callDef,
		executorFactory: func(ce *configv1.ContainerEnvironment) command.Executor {
			return mockExecutor
		},
	}

	req := &ExecutionRequest{
		ToolName:   "cmd-tool-json",
		ToolInputs: json.RawMessage(`{"input": "data"}`),
	}

	result, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "success", resMap["result"])
}

func TestCommandTool_Execute_Error_Mock(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("cmd-tool-error")

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("fail"),
	}

	callDef := &configv1.CommandLineCallDefinition{}

	mockExecutor := new(MockExecutor)

	// Setup mock to return error
	mockExecutor.On("Execute", mock.Anything, "fail", []string{}, "", mock.Anything).Return(
		io.NopCloser(bytes.NewBuffer(nil)),
		io.NopCloser(bytes.NewBuffer(nil)),
		(<-chan int)(nil),
		fmt.Errorf("command not found"),
	)

	tool := &CommandTool{
		tool:           toolProto,
		service:        service,
		callDefinition: callDef,
		executorFactory: func(ce *configv1.ContainerEnvironment) command.Executor {
			return mockExecutor
		},
	}

	req := &ExecutionRequest{
		ToolName:   "cmd-tool-error",
		ToolInputs: json.RawMessage(`{}`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute command")
}

func TestCommandTool_Execute_NonZeroExit_Mock(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("cmd-tool-fail")

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("failcmd"),
	}

	callDef := &configv1.CommandLineCallDefinition{}

	mockExecutor := new(MockExecutor)

	mockStdout := io.NopCloser(bytes.NewBufferString("some output"))
	mockStderr := io.NopCloser(bytes.NewBufferString("error occurred"))
	exitCodeChan := make(chan int, 1)
	exitCodeChan <- 1 // Non-zero exit code

	mockExecutor.On("Execute", mock.Anything, "failcmd", []string{}, "", mock.Anything).Return(
		mockStdout, mockStderr, (<-chan int)(exitCodeChan), nil,
	)

	tool := &CommandTool{
		tool:           toolProto,
		service:        service,
		callDefinition: callDef,
		executorFactory: func(ce *configv1.ContainerEnvironment) command.Executor {
			return mockExecutor
		},
	}

	req := &ExecutionRequest{
		ToolName:   "cmd-tool-fail",
		ToolInputs: json.RawMessage(`{}`),
	}

	result, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, resMap["return_code"])
}

func TestCommandTool_Execute_MalformedInput_Mock(t *testing.T) {
	tool := &CommandTool{}
	req := &ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: json.RawMessage(`{invalid-json`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

func TestCommandTool_MCPTool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("cmd-tool")
	toolProto.ServiceId = proto.String("test-service")
	toolProto.Description = proto.String("A command tool")

	tool := &CommandTool{
		tool: toolProto,
	}

	mcpTool := tool.MCPTool()
	assert.NotNil(t, mcpTool)
	assert.Equal(t, "test-service.cmd-tool", mcpTool.Name)
	assert.Equal(t, "A command tool", mcpTool.Description)

	// Call again to test Once
	mcpTool2 := tool.MCPTool()
	assert.Equal(t, mcpTool, mcpTool2)
}

func TestCommandTool_GetCacheConfig(t *testing.T) {
	toolProto := &v1.Tool{}
	callDef := &configv1.CommandLineCallDefinition{
		Cache: &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
		},
	}

	tool := &CommandTool{
		tool:           toolProto,
		callDefinition: callDef,
	}

	cacheConfig := tool.GetCacheConfig()
	assert.NotNil(t, cacheConfig)
	assert.True(t, cacheConfig.GetIsEnabled())

	toolNilDef := &CommandTool{
		tool: toolProto,
	}
	assert.Nil(t, toolNilDef.GetCacheConfig())
}

func TestLocalCommandTool_MCPTool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("local-cmd-tool")
	toolProto.ServiceId = proto.String("test-service")

	tool := &LocalCommandTool{
		tool: toolProto,
	}

	mcpTool := tool.MCPTool()
	assert.NotNil(t, mcpTool)
	assert.Equal(t, "test-service.local-cmd-tool", mcpTool.Name)
}

func TestLocalCommandTool_GetCacheConfig(t *testing.T) {
	toolProto := &v1.Tool{}
	callDef := &configv1.CommandLineCallDefinition{
		Cache: &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
		},
	}

	tool := &LocalCommandTool{
		tool:           toolProto,
		callDefinition: callDef,
	}

	assert.NotNil(t, tool.GetCacheConfig())
	assert.True(t, tool.GetCacheConfig().GetIsEnabled())

	toolNilDef := &LocalCommandTool{
		tool: toolProto,
	}
	assert.Nil(t, toolNilDef.GetCacheConfig())
}
