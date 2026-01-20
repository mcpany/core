// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/command"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

type MockExecutorForTest struct {
	mock.Mock
}

func (m *MockExecutorForTest) Execute(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	m.Called(ctx, cmd, args, workingDir, env)
	out := io.NopCloser(strings.NewReader("mock output"))
	err := io.NopCloser(strings.NewReader(""))
	ch := make(chan int, 1)
	ch <- 0
	close(ch)
	return out, err, ch, nil
}

func (m *MockExecutorForTest) ExecuteWithStdIO(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	argsMock := m.Called(ctx, cmd, args, workingDir, env)
	return argsMock.Get(0).(io.WriteCloser), argsMock.Get(1).(io.ReadCloser), argsMock.Get(2).(io.ReadCloser), argsMock.Get(3).(<-chan int), argsMock.Error(4)
}

func TestCommandTool_Execute_WithServiceArgs(t *testing.T) {
	mockExec := &MockExecutorForTest{}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Arguments: []string{"hello", "world"}, // Updated to Arguments
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"extra"},
	}

	toolDef := &v1.Tool{
		Name: proto.String("test-tool"),
	}

	tool := NewCommandTool(toolDef, service, callDef, nil, "call-id")

	// Inject mock executor factory
	ct, ok := tool.(*CommandTool)
	require.True(t, ok)
	ct.executorFactory = func(env *configv1.ContainerEnvironment) command.Executor {
		return mockExec
	}

	ctx := context.Background()
	req := &ExecutionRequest{
		ToolName: "test-tool",
		ToolInputs: []byte("{}"),
	}

	// Expectation: args should contain service args ("hello", "world") AND call args ("extra")
	// Note: We match ANY context here because the context might be wrapped
	mockExec.On("Execute", mock.Anything, "echo", []string{"hello", "world", "extra"}, "", mock.Anything).Return()

	_, err := tool.Execute(ctx, req)
	require.NoError(t, err)

	mockExec.AssertExpectations(t)
}
