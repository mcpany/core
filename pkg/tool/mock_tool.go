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

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
)

// MockTool is a mock implementation of the Tool interface for testing purposes.
type MockTool struct {
	ToolFunc           func() *v1.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool returns the protobuf definition of the mock tool.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
