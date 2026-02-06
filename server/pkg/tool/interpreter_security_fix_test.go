// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"io"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// sentinelMockExecutor implements command.Executor
type sentinelMockExecutor struct{}

func (m *sentinelMockExecutor) Execute(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	outR, outW := io.Pipe()
	outW.Close()
	errR, errW := io.Pipe()
	errW.Close()
	ch := make(chan int, 1)
	ch <- 0
	close(ch)
	return outR, errR, ch, nil
}

func (m *sentinelMockExecutor) ExecuteWithStdIO(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	return nil, nil, nil, nil, nil
}

func TestInterpreterInjectionBypass(t *testing.T) {
	toolDef := v1.Tool_builder{
		Name: proto.String("python-tool"),
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{input1}}', '{{input2}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input1"),
				}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input2"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	cmdTool := NewCommandTool(toolDef, service, callDef, nil, "call-1")
	ct, ok := cmdTool.(*CommandTool)
	assert.True(t, ok)
	ct.executorFactory = func(env *configv1.ContainerEnvironment) command.Executor {
		return &sentinelMockExecutor{}
	}

	t.Run("Vulnerability_Blocked", func(t *testing.T) {
		// Attack Scenario: input1 = "\" (Backslash) -> Escapes closing quote
		input1 := "\\"
		input2 := `, open("/etc/passwd")) #`

		// Construct JSON inputs
		escapedInput1 := strings.ReplaceAll(input1, `\`, `\\`)
		escapedInput2 := strings.ReplaceAll(input2, `"`, `\"`)
		jsonInputs := `{"input1": "` + escapedInput1 + `", "input2": "` + escapedInput2 + `"}`

		req := &ExecutionRequest{
			ToolName:   "python-tool",
			ToolInputs: []byte(jsonInputs),
		}

		_, err := cmdTool.Execute(context.Background(), req)

		if err == nil {
			t.Fatal("Vulnerability Confirmed: Backslash injection allowed! The fix is not working.")
		} else {
			t.Logf("Execution failed: %v", err)
			if strings.Contains(err.Error(), "injection detected") {
				t.Log("Secure: Injection detected.")
			} else {
				t.Fatalf("Execution failed with unexpected error: %v", err)
			}
		}
	})

	t.Run("Safe_Middle_Backslash_Allowed", func(t *testing.T) {
		// Safe: "abc\def" -> Python 'abc\def'. Not an escape sequence for '
		input1 := `abc\def`

		escapedInput1 := strings.ReplaceAll(input1, `\`, `\\`)
		jsonInputs := `{"input1": "` + escapedInput1 + `", "input2": "safe"}`

		req := &ExecutionRequest{
			ToolName:   "python-tool",
			ToolInputs: []byte(jsonInputs),
		}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.NoError(t, err, "Safe backslash in middle should be allowed")
	})

	t.Run("Safe_Double_Backslash_Allowed", func(t *testing.T) {
		// Safe: "abc\\" -> Python 'abc\\' (literal backslash). Quote is safe.
		input1 := `abc\\`

		escapedInput1 := strings.ReplaceAll(input1, `\`, `\\`)
		jsonInputs := `{"input1": "` + escapedInput1 + `", "input2": "safe"}`

		req := &ExecutionRequest{
			ToolName:   "python-tool",
			ToolInputs: []byte(jsonInputs),
		}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.NoError(t, err, "Double backslash at end should be allowed")
	})
}
