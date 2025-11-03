/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool

import (
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestToolManager_AddTool_EmptyServiceID(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("")
	toolProto.SetName("test-tool")
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.Error(t, err, "Should return an error for a tool with an empty service ID")
	assert.Equal(t, "tool service ID cannot be empty", err.Error())
}

func TestToolManager_SetMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	mockProvider := new(MockMCPServerProvider)
	mockProvider.On("Server").Return((*mcp.Server)(nil))
	tm.SetMCPServer(mockProvider)
	assert.Equal(t, mockProvider, tm.mcpServer, "MCPServerProvider should be set")
}
