
// Copyright 2024 Author of MCP any
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

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalCommandTool(t *testing.T) {
	tool := NewLocalCommandTool(
		&v1.Tool{},
		&configv1.CommandLineUpstreamService{},
		&configv1.CommandLineCallDefinition{},
	)
	assert.NotNil(t, tool)
}

func TestLocalCommandTool_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		svc := &configv1.CommandLineUpstreamService{}
		svc.SetCommand("echo")
		tool := NewLocalCommandTool(
			&v1.Tool{},
			svc,
			&configv1.CommandLineCallDefinition{},
		)

		req := &ExecutionRequest{
			ToolInputs: json.RawMessage(`{"args": ["hello", "world"]}`),
		}

		res, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, resMap["stdout"], "hello world")
	})

	t.Run("command not found", func(t *testing.T) {
		svc := &configv1.CommandLineUpstreamService{}
		svc.SetCommand("non-existent-command-12345")
		tool := NewLocalCommandTool(
			&v1.Tool{},
			svc,
			&configv1.CommandLineCallDefinition{},
		)

		req := &ExecutionRequest{
			ToolInputs: json.RawMessage(`{}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("with args in definition", func(t *testing.T) {
		svc := &configv1.CommandLineUpstreamService{}
		svc.SetCommand("echo")
		tool := NewLocalCommandTool(
			&v1.Tool{},
			svc,
			&configv1.CommandLineCallDefinition{
				Args: []string{"hello"},
			},
		)

		req := &ExecutionRequest{
			ToolInputs: json.RawMessage(`{"args": ["world"]}`),
		}

		res, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, resMap["stdout"], "world")
	})

	t.Run("malformed tool inputs", func(t *testing.T) {
		svc := &configv1.CommandLineUpstreamService{}
		svc.SetCommand("echo")
		tool := NewLocalCommandTool(
			&v1.Tool{},
			svc,
			&configv1.CommandLineCallDefinition{},
		)

		req := &ExecutionRequest{
			ToolInputs: json.RawMessage(`{"args": "not-an-array"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestLocalCommandTool_Tool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	svc := &configv1.CommandLineUpstreamService{}
	svc.SetCommand("echo")
	tool := NewLocalCommandTool(
		toolProto,
		svc,
		&configv1.CommandLineCallDefinition{},
	)
	assert.Equal(t, toolProto, tool.Tool())
}

func TestLocalCommandTool_GetCacheConfig(t *testing.T) {
	cacheConfig := &configv1.CacheConfig{}
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetCache(cacheConfig)
	tool := NewLocalCommandTool(
		&v1.Tool{},
		&configv1.CommandLineUpstreamService{},
		callDef,
	)
	assert.Equal(t, cacheConfig, tool.GetCacheConfig())
}
