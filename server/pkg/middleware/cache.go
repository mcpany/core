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
	"github.com/mcpany/core/pkg/ai/embeddings"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache       *cache.Cache[any]
	toolManager tool.ManagerInterface
	vectorStore *VectorStore
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:       cacheManager,
		toolManager: toolManager,
		vectorStore: NewVectorStore(),
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

	// Inject CacheControl if not present
	var cacheControl *tool.CacheControl
	if cc, ok := tool.GetCacheControl(ctx); ok {
		cacheControl = cc
	} else {
		cacheControl = &tool.CacheControl{Action: tool.ActionAllow}
		ctx = tool.NewContextWithCacheControl(ctx, cacheControl)
	}

	cacheKey := m.getCacheKey(req)
	var semanticConfig *configv1.SemanticCacheConfig
	if cacheConfig != nil {
		semanticConfig = cacheConfig.GetSemantic()
	}

	// Check cache ONLY if action is not DeleteCache
	if cacheControl.Action != tool.ActionDeleteCache {
		// 1. Check Exact Match
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			// Found in cache
			metrics.IncrCounterWithLabels([]string{"cache", "hits"}, 1, labels)
			return cached, nil
		}

		// 2. Check Semantic Match
		if semanticConfig != nil && semanticConfig.GetIsEnabled() {
			embedder, err := embeddings.NewEmbedder(semanticConfig.GetEmbeddingProvider(), semanticConfig.GetEmbeddingModel())
			if err == nil {
				// Convert inputs to text.
				inputText := string(req.ToolInputs)
				vecs, err := embedder.Embed(ctx, []string{inputText})
				if err == nil && len(vecs) > 0 {
					if result, found := m.vectorStore.Search(req.ToolName, vecs[0], semanticConfig.GetSimilarityThreshold()); found {
						metrics.IncrCounterWithLabels([]string{"cache", "semantic_hits"}, 1, labels)
						return result, nil
					}
				}
			}
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
		// Also clear semantic cache? VectorStore doesn't support targeted delete by key yet,
		// and we don't know the exact vector.
		// For now we skip semantic delete.
		return result, nil
	}

	ttl := cacheConfig.GetTtl().AsDuration()

	if err := m.cache.Set(ctx, cacheKey, result, store.WithExpiration(ttl)); err != nil {
		metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
		logging.GetLogger().Error("Failed to set cache", "error", err, "tool", toolName)
	}

	// Store Semantic
	if semanticConfig != nil && semanticConfig.GetIsEnabled() {
		embedder, err := embeddings.NewEmbedder(semanticConfig.GetEmbeddingProvider(), semanticConfig.GetEmbeddingModel())
		if err == nil {
			inputText := string(req.ToolInputs)
			vecs, err := embedder.Embed(ctx, []string{inputText})
			if err == nil && len(vecs) > 0 {
				m.vectorStore.Add(req.ToolName, vecs[0], result, ttl)
			}
		}
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
	return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
}

// Clear clears the cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	m.vectorStore.Clear()
	return m.cache.Clear(ctx)
}
