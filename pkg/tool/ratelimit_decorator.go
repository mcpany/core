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

	"github.com/mcpany/core/pkg/policy/ratelimit"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimitedTool is a decorator that adds rate limiting to a Tool.
type RateLimitedTool struct {
	tool    Tool
	limiter ratelimit.Limiter
}

// NewRateLimitedTool creates a new RateLimitedTool.
func NewRateLimitedTool(tool Tool, limiter ratelimit.Limiter) *RateLimitedTool {
	return &RateLimitedTool{
		tool:    tool,
		limiter: limiter,
	}
}

// Tool returns the underlying tool's definition.
func (t *RateLimitedTool) Tool() *v1.Tool {
	return t.tool.Tool()
}

// Execute checks the rate limiter before executing the tool.
func (t *RateLimitedTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if !t.limiter.Allow() {
		return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
	}
	return t.tool.Execute(ctx, req)
}

// GetCacheConfig returns the underlying tool's cache configuration.
func (t *RateLimitedTool) GetCacheConfig() *configv1.CacheConfig {
	return t.tool.GetCacheConfig()
}
