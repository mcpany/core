// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache          *cache.Cache[any]
	toolManager    tool.ManagerInterface
	semanticCaches sync.Map
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:       cacheManager,
		toolManager: toolManager,
	}
}

// Execute executes the caching middleware.
func (m *CachingMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := tool.GetFromContext(ctx)
	if !ok {
		// Tool not in context, try to look it up
		var found bool
		t, found = m.toolManager.GetTool(req.ToolName)
		if !found {
			logging.GetLogger().Debug("Tool not found in manager", "tool", req.ToolName)
			return next(ctx, req)
		}
	}

	logging.GetLogger().Debug("CachingMiddleware Execute", "tool", t.Tool().GetName())
	cacheConfig := m.getCacheConfig(t)
	if cacheConfig == nil || !cacheConfig.GetIsEnabled() {
		logging.GetLogger().Debug("Caching disabled or config nil", "tool", t.Tool().GetName())
		return next(ctx, req)
	}

	// Extract service ID and Tool Name for metrics
	serviceID := t.Tool().GetServiceId()
	toolName := t.Tool().GetName()
	labels := []metrics.Label{
		{Name: "service", Value: serviceID},
		{Name: "tool", Value: toolName},
	}

	if cacheConfig.GetStrategy() == "semantic" {
		return m.executeSemantic(ctx, t, req, next, cacheConfig, labels)
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
		metrics.IncrCounterWithLabels([]string{"cache", "skips"}, 1, labels)
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check CacheControl
	if cacheControl.Action == tool.ActionDeleteCache {
		if err := m.cache.Delete(ctx, cacheKey); err != nil {
			metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
			logging.GetLogger().Error("Failed to delete cache", "error", err, "tool", toolName)
		}
		return result, nil
	}

	if err := m.cache.Set(ctx, cacheKey, result, store.WithExpiration(cacheConfig.GetTtl().AsDuration())); err != nil {
		metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
		logging.GetLogger().Error("Failed to set cache", "error", err, "tool", toolName)
	}
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
	// Normalize ToolInputs if they are JSON
	var normalizedInputs []byte
	if len(req.ToolInputs) > 0 {
		var input any
		// We use standard json.Unmarshal which sorts map keys when Marshaling back.
		if err := json.Unmarshal(req.ToolInputs, &input); err == nil {
			if marshaled, err := json.Marshal(input); err == nil {
				normalizedInputs = marshaled
			}
		}
	}

	// Fallback to raw bytes if unmarshal fails or empty
	if normalizedInputs == nil {
		normalizedInputs = req.ToolInputs
	}

	return fmt.Sprintf("%s:%s", req.ToolName, normalizedInputs)
}

func (m *CachingMiddleware) executeSemantic(ctx context.Context, t tool.Tool, req *tool.ExecutionRequest, next tool.ExecutionFunc, config *configv1.CacheConfig, labels []metrics.Label) (any, error) {
	serviceID := t.Tool().GetServiceId()

	var semCache *SemanticCache
	if val, ok := m.semanticCaches.Load(serviceID); ok {
		semCache = val.(*SemanticCache)
	} else {
		semConfig := config.GetSemanticConfig()
		if semConfig == nil {
			logging.GetLogger().Warn("Semantic strategy selected but no semantic config found", "tool", t.Tool().GetName())
			// Fallback to exact match logic (basically continuing is hard without refactoring, so we just skip cache here)
			return next(ctx, req)
		}

		var provider EmbeddingProvider
		if semConfig.GetProvider() == "openai" {
			provider = NewOpenAIEmbeddingProvider(semConfig.GetApiKey(), semConfig.GetModel())
		} else {
			logging.GetLogger().Warn("Unknown embedding provider", "provider", semConfig.GetProvider())
			return next(ctx, req)
		}

		semCache = NewSemanticCache(provider, semConfig.GetSimilarityThreshold())
		m.semanticCaches.Store(serviceID, semCache)
	}

	inputBytes, err := json.Marshal(req.ToolInputs)
	if err != nil {
		logging.GetLogger().Error("Failed to marshal tool inputs for semantic cache", "error", err)
		return next(ctx, req)
	}
	inputStr := string(inputBytes)

	// Check cache
	cached, hit, err := semCache.Get(ctx, req.ToolName, inputStr)
	if err != nil {
		logging.GetLogger().Error("Semantic cache error", "error", err)
		return next(ctx, req)
	}

	if hit {
		metrics.IncrCounterWithLabels([]string{"cache", "hits"}, 1, labels)
		return cached, nil
	}

	metrics.IncrCounterWithLabels([]string{"cache", "misses"}, 1, labels)

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	// Set cache
	if err := semCache.Set(ctx, req.ToolName, inputStr, result, config.GetTtl().AsDuration()); err != nil {
		logging.GetLogger().Error("Failed to set semantic cache", "error", err)
	}

	return result, nil
}

// Clear clears the cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	// Also clear semantic caches
	m.semanticCaches.Range(func(key, value any) bool {
		m.semanticCaches.Delete(key)
		return true
	})
	return m.cache.Clear(ctx)
}
