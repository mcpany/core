// Copyright 2025 Author(s) of MCP Any
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
//
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Execute(t *testing.T) {
	tool := &v1.Tool{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("echo")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{}

	localTool := NewLocalCommandTool(tool, service, callDef)

	req := &ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"hello", "world"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello world\n", resultMap["stdout"])

	assert.NotNil(t, localTool.Tool())
	assert.Equal(t, tool, localTool.Tool())
	assert.Nil(t, localTool.GetCacheConfig())
}
