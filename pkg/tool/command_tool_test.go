
// Copyright 2024 Author(s) of MCP any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func newCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = &configv1.CommandLineCallDefinition{}
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String(command),
	}).Build()
	return tool.NewCommandTool(
		&v1.Tool{},
		service,
		callDef,
	)
}

func newJSONCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = &configv1.CommandLineCallDefinition{}
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command:               proto.String(command),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}).Build()
	return tool.NewCommandTool(
		&v1.Tool{},
		service,
		callDef,
	)
}

func TestCommandTool_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		cmdTool := newCommandTool("/usr/bin/env", nil)
		inputData := map[string]interface{}{"args": []string{"echo", "hello world"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "hello world\n", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "hello world\n", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])
		assert.Equal(t, 0, resultMap["return_code"])
	})

	t.Run("command not found", func(t *testing.T) {
		cmdTool := newCommandTool("this-command-does-not-exist", nil)
		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("execution with environment variables", func(t *testing.T) {
		cmdTool := newCommandTool("/usr/bin/env", nil)
		inputData := map[string]interface{}{
			"args":   []string{"bash", "-c", "echo $MY_VAR"},
			"MY_VAR": "hello from env",
		}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "hello from env\n", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "hello from env\n", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])
		assert.Equal(t, 0, resultMap["return_code"])
	})

	t.Run("non-zero exit code", func(t *testing.T) {
		cmdTool := newCommandTool("/usr/bin/env", nil)
		inputData := map[string]interface{}{"args": []string{"bash", "-c", "exit 1"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusError, resultMap["status"])
		assert.Equal(t, 1, resultMap["return_code"])
	})

	t.Run("malformed tool inputs", func(t *testing.T) {
		cmdTool := newCommandTool("echo", nil)
		inputs := json.RawMessage(`{"args": "not-an-array"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("json communication protocol", func(t *testing.T) {
		cmdTool := newJSONCommandTool("./testdata/jsonecho/jsonecho", nil)
		inputData := map[string]interface{}{"foo": "bar"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "bar", resultMap["foo"])
	})
}

func TestCommandTool_GetCacheConfig(t *testing.T) {
	cacheConfig := &configv1.CacheConfig{}
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetCache(cacheConfig)
	cmdTool := newCommandTool("echo", callDef)
	assert.Equal(t, cacheConfig, cmdTool.GetCacheConfig())
}

func TestCommandTool_Tool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}).Build()
	cmdTool := tool.NewCommandTool(toolProto, service, &configv1.CommandLineCallDefinition{})
	assert.Equal(t, toolProto, cmdTool.Tool())
}

// createTempScript creates a temporary executable script for testing.
func createTempScript(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp("", "test-script-*.sh")
	require.NoError(t, err)
	_, err = file.WriteString(content)
	require.NoError(t, err)
	err = file.Chmod(0755)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	return file.Name()
}

func TestCommandTool_Execute_WithTimeout(t *testing.T) {
	t.Run("timeout exceeded", func(t *testing.T) {
		// Create a temporary script that sleeps for a long time
		scriptPath := createTempScript(t, "#!/bin/bash\nsleep 10")
		defer os.Remove(scriptPath)

		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{
				Command: proto.String(scriptPath),
				Timeout: durationpb.New(10 * time.Millisecond),
			}).Build(),
			nil,
		)

		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		result, err := cmdTool.Execute(context.Background(), req)

		require.NoError(t, err, "Execute should not return an error, even on timeout")

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		assert.Equal(t, consts.CommandStatusTimeout, resultMap["status"], "Status should be timeout")
	})
}

func TestCommandTool_Execute_WithInput(t *testing.T) {
	t.Run("stdin provided", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{Command: proto.String("cat")}).Build(),
			nil,
		)

		inputData := map[string]interface{}{"stdin": "hello from stdin"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from stdin", resultMap["stdout"])
	})
}

func TestCommandTool_Execute_WithArgs(t *testing.T) {
	t.Run("args from tool definition", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{
				Command: proto.String("echo"),
			}).Build(),
			&configv1.CommandLineCallDefinition{
				Args: []string{"hello", "from", "definition"},
			},
		)

		req := &tool.ExecutionRequest{ToolInputs: []byte(`{}`)}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from definition\n", resultMap["stdout"])
	})

	t.Run("args from request", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}).Build(),
			nil,
		)

		inputData := map[string]interface{}{"args": []string{"hello", "from", "request"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from request\n", resultMap["stdout"])
	})

	t.Run("args from both, request taking precedence", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{
				Command: proto.String("echo"),
			}).Build(),
			&configv1.CommandLineCallDefinition{
				Args: []string{"this is ignored"},
			},
		)

		inputData := map[string]interface{}{"args": []string{"hello", "from", "request"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from request\n", resultMap["stdout"])
	})
}

func TestCommandTool_Execute_JSONCommunication(t *testing.T) {
	scriptContent := `#!/bin/bash
read input
echo "$input"
`
	scriptPath := createTempScript(t, scriptContent)
	defer os.Remove(scriptPath)

	t.Run("successful json communication", func(t *testing.T) {
		cmdTool := newJSONCommandTool(scriptPath, nil)

		inputData := map[string]interface{}{"message": "hello"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		expectedOutput := map[string]interface{}{"message": "hello"}
		assert.Equal(t, expectedOutput, resultMap)
	})

	t.Run("invalid output is not parsed", func(t *testing.T) {
		// This script outputs a non-JSON string
		nonJSONScript := createTempScript(t, "#!/bin/bash\necho 'not json'")
		defer os.Remove(nonJSONScript)

		cmdTool := newJSONCommandTool(nonJSONScript, nil)

		req := &tool.ExecutionRequest{ToolInputs: []byte(`{}`)}
		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestCommandTool_Execute_WithWorkingDirectory(t *testing.T) {
	// Create a temporary directory and a file inside it
	tmpDir, err := os.MkdirTemp("", "test-wd")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(tmpDir+"/testfile.txt", []byte("hello"), 0644)
	require.NoError(t, err)

	t.Run("successful execution in specified working directory", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{
				Command:          proto.String("cat"),
				WorkingDirectory: proto.String(tmpDir),
			}).Build(),
			&configv1.CommandLineCallDefinition{
				Args: []string{"testfile.txt"},
			},
		)

		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello", resultMap["stdout"])
	})

	t.Run("error on non-existent working directory", func(t *testing.T) {
		cmdTool := tool.NewCommandTool(
			nil,
			(&configv1.CommandLineUpstreamService_builder{
				Command:          proto.String("pwd"),
				WorkingDirectory: proto.String("/path/to/non/existent/dir"),
			}).Build(),
			nil,
		)

		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		_, err = cmdTool.Execute(context.Background(), req)
		assert.Error(t, err, "Expected an error for non-existent working directory")
	})
}

func TestCommandTool_Execute_ContextCancellation(t *testing.T) {
	scriptPath := createTempScript(t, "#!/bin/bash\nsleep 5")
	defer os.Remove(scriptPath)

	cmdTool := tool.NewCommandTool(
		nil,
		(&configv1.CommandLineUpstreamService_builder{Command: proto.String(scriptPath)}).Build(),
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}

	var result interface{}
	var err error
	done := make(chan struct{})

	go func() {
		result, err = cmdTool.Execute(ctx, req)
		close(done)
	}()

	// Give the command a moment to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for the Execute function to return
	<-done

	require.NoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Check the status and error message
	assert.Equal(t, consts.CommandStatusError, resultMap["status"])
}
