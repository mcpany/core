// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache         *cache.Cache[any]
	defaultConfig *configv1.CacheConfig
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(config *configv1.CacheConfig) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:         cacheManager,
		defaultConfig: config,
	}
}

// Execute executes the caching middleware.
func (m *CachingMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// Look for tool in context to check for overrides
	t, ok := tool.GetFromContext(ctx)

	cacheConfig := m.getCacheConfig(t, ok)
	if cacheConfig == nil || !cacheConfig.GetIsEnabled() {
		return next(ctx, req)
	}

	// Extract service ID and Tool Name for metrics
	var serviceID, toolName string
	if ok {
		serviceID = t.Tool().GetServiceId()
		toolName = t.Tool().GetName()
	} else {
		// Fallback if tool not in context (shouldn't happen in proper chain)
		toolName = req.ToolName
		serviceID = "unknown"
	}

	labels := []metrics.Label{
		{Name: "service", Value: serviceID},
		{Name: "tool", Value: toolName},
	}

	// Inject CacheControl if not present
	var cacheControl *tool.CacheControl
	if cc, ok := tool.GetCacheControl(ctx); ok {
		cacheControl = cc
	} else {
		cacheControl = &tool.CacheControl{Action: tool.ActionAllow}
		ctx = tool.NewContextWithCacheControl(ctx, cacheControl)
	}

	cacheKey := m.getCacheKey(req)

	// Check cache ONLY if action is not DeleteCache
	if cacheControl.Action != tool.ActionDeleteCache {
		// If normal allow (0), check cache.
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			// Found in cache
			metrics.IncrCounterWithLabels([]string{"cache", "hits"}, 1, labels)
			return cached, nil
		}
		// Not found in cache
		metrics.IncrCounterWithLabels([]string{"cache", "misses"}, 1, labels)
	} else {
		// If DeleteCache, we skip cache lookup (force miss)
		metrics.IncrCounterWithLabels([]string{"cache", "skips"}, 1, labels) // Optional: track skips
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check CacheControl
	if cacheControl.Action == tool.ActionDeleteCache {
		if err := m.cache.Delete(ctx, cacheKey); err != nil {
			_ = err
		}
		return result, nil
	}

	if err := m.cache.Set(ctx, cacheKey, result, store.WithExpiration(cacheConfig.GetTtl().AsDuration())); err != nil {
		_ = err // explicit ignore
	}
	return result, nil
}

func (m *CachingMiddleware) getCacheConfig(t tool.Tool, found bool) *configv1.CacheConfig {
	if found {
		if callCacheConfig := t.GetCacheConfig(); callCacheConfig != nil {
			return callCacheConfig
		}
	}
	return m.defaultConfig
}

func (m *CachingMiddleware) getCacheKey(req *tool.ExecutionRequest) string {
	return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
}

// Clear clears the cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	return m.cache.Clear(ctx)
}
