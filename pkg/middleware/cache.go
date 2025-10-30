/*
 * Copyright 2025 Author(s) of MCP-XY
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

package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpxy/core/pkg/common/clock"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/patrickmn/go-cache"
)

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache       *cache.Cache
	toolManager tool.ToolManagerInterface
	clock       clock.Clock
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ToolManagerInterface, clock clock.Clock) *CachingMiddleware {
	return &CachingMiddleware{
		cache:       cache.New(5*time.Minute, 10*time.Minute),
		toolManager: toolManager,
		clock:       clock,
	}
}

// Execute executes the caching middleware.
func (m *CachingMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ToolExecutionFunc) (any, error) {
	t, ok := tool.GetFromContext(ctx)
	if !ok {
		return next(ctx, req)
	}

	cacheConfig := m.getCacheConfig(t)
	if cacheConfig == nil || !cacheConfig.GetIsEnabled() {
		return next(ctx, req)
	}

	cacheKey := m.getCacheKey(req)
	if cached, found := m.cache.Get(cacheKey); found {
		if entry, ok := cached.(cacheEntry); ok {
			if m.clock.Now().Before(entry.expiresAt) {
				return entry.data, nil
			}
		}
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	entry := cacheEntry{
		data:      result,
		expiresAt: m.clock.Now().Add(cacheConfig.GetTtl().AsDuration()),
	}
	m.cache.Set(cacheKey, entry, cache.NoExpiration)
	return result, nil
}

func (m *CachingMiddleware) getCacheConfig(t tool.Tool) *configv1.CacheConfig {
	if callCacheConfig := t.GetCacheConfig(); callCacheConfig != nil {
		return callCacheConfig
	}

	serviceInfo, ok := m.toolManager.GetServiceInfo(t.Tool().GetServiceId())
	if !ok {
		return nil
	}

	return serviceInfo.Config.GetCache()
}

func (m *CachingMiddleware) getCacheKey(req *tool.ExecutionRequest) string {
	return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
}
