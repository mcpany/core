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
	"context"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolExecutionMiddleware is a mock implementation of the ToolExecutionMiddleware interface.
type MockToolExecutionMiddleware struct {
	mock.Mock
}

func (m *MockToolExecutionMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
	args := m.Called(ctx, req, next)
	return args.Get(0), args.Error(1)
}

func TestToolManager_AddMiddleware(t *testing.T) {
	tm := NewToolManager(nil)
	middleware := new(MockToolExecutionMiddleware)
	tm.AddMiddleware(middleware)
	assert.Len(t, tm.middlewares, 1, "Should have one middleware")
	assert.Equal(t, middleware, tm.middlewares[0], "The middleware should be the one that was added")
}

func TestToolManager_SetMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	mockProvider := new(MockMCPServerProvider)
	tm.SetMCPServer(mockProvider)
	assert.NotNil(t, tm.mcpServer, "mcpServer should be set")
}

func TestToolManager_AddToolWithMCPServer_InvalidToolName(t *testing.T) {
	b := bus.NewBusProvider()
	tm := NewToolManager(b)

	toolProto := v1.Tool_builder{
		ServiceId: proto.String("test-service"),
		Name:      proto.String(""), // Invalid name
	}.Build()

	mockTool := new(MockTool)
	mockTool.On("Tool").Return(toolProto)

	mockServer := NewMockMCPToolServer()
	mockProvider := new(MockMCPServerProvider)
	mockProvider.On("Server").Return(mockServer.Server)

	tm.SetMCPServer(mockProvider)

	err := tm.AddTool(mockTool)
	assert.Error(t, err, "Should return an error for an invalid tool name")
}
