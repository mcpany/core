// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/mcpany/core/pkg/embedder"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/vectorstore"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache       *cache.Cache[any]
	toolManager tool.ManagerInterface
	// Semantic caching components
	vectorStore vectorstore.Store
	embedder    embedder.Embedder
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:       cacheManager,
		toolManager: toolManager,
		vectorStore: vectorstore.NewSimpleStore(10000), // Cap at 10k entries
		embedder:    embedder.NewBagOfWordsEmbedder(128), // 128-dim embeddings
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

	// Determine strategy
	strategy := cacheConfig.GetType()
	if strategy == configv1.CacheConfig_STRATEGY_UNSPECIFIED {
		// Fallback to deprecated strategy string check
		strategy = configv1.CacheConfig_STRATEGY_EXACT
	}

	// Handle Semantic Caching
	if strategy == configv1.CacheConfig_STRATEGY_SEMANTIC {
		return m.executeSemantic(ctx, req, next, cacheConfig, cacheControl, labels)
	}

	// Handle Exact Caching (Default)
	return m.executeExact(ctx, req, next, cacheConfig, cacheControl, labels)
}

func (m *CachingMiddleware) executeExact(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc, cacheConfig *configv1.CacheConfig, cacheControl *tool.CacheControl, labels []metrics.Label) (any, error) {
	cacheKey := m.getCacheKey(req)

	// Check cache ONLY if action is not DeleteCache
	if cacheControl.Action != tool.ActionDeleteCache {
		// If normal allow (0), check cache.
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			// Found in cache
			metrics.IncrCounterWithLabels([]string{"cache", "exact", "hits"}, 1, labels)
			return cached, nil
		}
		// Not found in cache
		metrics.IncrCounterWithLabels([]string{"cache", "exact", "misses"}, 1, labels)
	} else {
		// If DeleteCache, we skip cache lookup (force miss)
		metrics.IncrCounterWithLabels([]string{"cache", "exact", "skips"}, 1, labels)
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check CacheControl
	if cacheControl.Action == tool.ActionDeleteCache {
		if err := m.cache.Delete(ctx, cacheKey); err != nil {
			metrics.IncrCounterWithLabels([]string{"cache", "exact", "errors"}, 1, labels)
			logging.GetLogger().Error("Failed to delete cache", "error", err, "tool", req.ToolName)
		}
		return result, nil
	}

	if err := m.cache.Set(ctx, cacheKey, result, store.WithExpiration(cacheConfig.GetTtl().AsDuration())); err != nil {
		metrics.IncrCounterWithLabels([]string{"cache", "exact", "errors"}, 1, labels)
		logging.GetLogger().Error("Failed to set cache", "error", err, "tool", req.ToolName)
	}
	return result, nil
}

func (m *CachingMiddleware) executeSemantic(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc, cacheConfig *configv1.CacheConfig, cacheControl *tool.CacheControl, labels []metrics.Label) (any, error) {
	inputText := string(req.ToolInputs)

	// Check provider (logging only for MVP)
	if sc := cacheConfig.GetSemanticConfig(); sc != nil {
		provider := sc.GetEmbeddingProvider()
		if provider != "" && provider != "bag-of-words" && provider != "mock" {
			logging.GetLogger().Warn("Unsupported embedding provider, falling back to bag-of-words", "provider", provider)
		}
	}

	embedding, err := m.embedder.Embed(inputText)
	if err != nil {
		logging.GetLogger().Error("Failed to embed input", "error", err)
		return next(ctx, req)
	}

	threshold := float32(0.9) // Default
	if sc := cacheConfig.GetSemanticConfig(); sc != nil && sc.GetSimilarityThreshold() > 0 {
		threshold = sc.GetSimilarityThreshold()
	}

	if cacheControl.Action != tool.ActionDeleteCache {
		results, err := m.vectorStore.Search(embedding, 1, threshold)
		if err == nil && len(results) > 0 {
			// Hit
			metrics.IncrCounterWithLabels([]string{"cache", "semantic", "hits"}, 1, labels)
			logging.GetLogger().Debug("Semantic cache hit", "similarity", results[0].Similarity)
			return results[0].Data, nil
		}
		metrics.IncrCounterWithLabels([]string{"cache", "semantic", "misses"}, 1, labels)
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	if cacheControl.Action == tool.ActionDeleteCache {
		// Semantic cache doesn't support fine-grained deletion yet.
		return result, nil
	}

	// Store in vector store with TTL
	ttl := cacheConfig.GetTtl().AsDuration()
	if ttl == 0 {
		ttl = 1 * time.Hour // Default semantic cache TTL
	}

	if err := m.vectorStore.Add(embedding, result, ttl); err != nil {
		metrics.IncrCounterWithLabels([]string{"cache", "semantic", "errors"}, 1, labels)
		logging.GetLogger().Error("Failed to add to semantic cache", "error", err)
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

// Clear clears the cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	m.vectorStore.Clear()
	return m.cache.Clear(ctx)
}
