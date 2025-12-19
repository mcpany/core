// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/command"
	"github.com/mcpany/core/pkg/consts"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockExecutor struct {
	CapturedCommand    string
	CapturedArgs       []string
	CapturedWorkingDir string
	CapturedEnv        []string

	Stdout     string
	Stderr     string
	ExitCode   int
	ReturnErr  error
}

func (m *mockExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	m.CapturedCommand = command
	m.CapturedArgs = args
	m.CapturedWorkingDir = workingDir
	m.CapturedEnv = env

	if m.ReturnErr != nil {
		return nil, nil, nil, m.ReturnErr
	}

	outR := io.NopCloser(strings.NewReader(m.Stdout))
	errR := io.NopCloser(strings.NewReader(m.Stderr))

	ch := make(chan int, 1)
	ch <- m.ExitCode
	close(ch)

	return outR, errR, ch, nil
}

func (m *mockExecutor) ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	// Not implemented for this test
	return nil, nil, nil, nil, nil
}

func TestCommandTool_Execute_WithContainerEnvironment(t *testing.T) {
	mockExec := &mockExecutor{
		Stdout: "container output",
		ExitCode: 0,
	}

	toolDef := &v1.Tool{Name: proto.String("container-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("run-in-container"),
		ContainerEnvironment: &configv1.ContainerEnvironment{
			Image: proto.String("busybox"),
			Env: map[string]*configv1.SecretValue{
				"CONTAINER_VAR": {
					Value: &configv1.SecretValue_PlainText{PlainText: "val"},
				},
			},
		},
	}
	callDef := &configv1.CommandLineCallDefinition{}

	// Manually constructing CommandTool to inject factory
	cmdTool := &CommandTool{
		tool:           toolDef,
		service:        service,
		callDefinition: callDef,
		callID:         "call-id",
		executorFactory: func(env *configv1.ContainerEnvironment) command.Executor {
			assert.Equal(t, "busybox", env.GetImage())
			return mockExec
		},
	}

	req := &ExecutionRequest{
		ToolName:   "container-tool",
		ToolInputs: []byte("{}"),
	}

	result, err := cmdTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "container output", resultMap["stdout"])
	assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])

	assert.Equal(t, "run-in-container", mockExec.CapturedCommand)

	// Check if CONTAINER_VAR is in environment
	found := false
	for _, e := range mockExec.CapturedEnv {
		if e == "CONTAINER_VAR=val" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected CONTAINER_VAR to be passed to executor")
}
