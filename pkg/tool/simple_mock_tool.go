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

	"github.com/mcpany/core/proto/config/v1"
	mcpv1 "github.com/mcpany/core/proto/mcp_router/v1"
)

// SimpleMockTool is a simplified mock of the Tool interface for testing.
type SimpleMockTool struct {
	ExecuteFunc    func(ctx context.Context, req *ExecutionRequest) (interface{}, error)
	ToolFunc       func() *mcpv1.Tool
	GetCacheConfigFunc func() *v1.CacheConfig
}

func (m *SimpleMockTool) Execute(ctx context.Context, req *ExecutionRequest) (interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

func (m *SimpleMockTool) Tool() *mcpv1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return nil
}

func (m *SimpleMockTool) GetCacheConfig() *v1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}

// SimpleMockToolExecutionMiddleware is a simplified mock of the ToolExecutionMiddleware interface for testing.
type SimpleMockToolExecutionMiddleware struct {
	ExecuteFunc func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (interface{}, error)
}

func (m *SimpleMockToolExecutionMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req, next)
	}
	return next(ctx, req)
}
