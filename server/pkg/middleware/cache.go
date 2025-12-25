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
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// ProviderFactory is a function that creates an EmbeddingProvider.
type ProviderFactory func(config *configv1.SemanticCacheConfig, apiKey string) (EmbeddingProvider, error)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	cache           *cache.Cache[any]
	toolManager     tool.ManagerInterface
	semanticCaches  sync.Map
	providerFactory ProviderFactory
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:       cacheManager,
		toolManager: toolManager,
		providerFactory: func(conf *configv1.SemanticCacheConfig, apiKey string) (EmbeddingProvider, error) {
			// Check OneOf provider_config first
			if conf.GetOpenai() != nil {
				openaiConf := conf.GetOpenai()
				key := apiKey
				if key == "" {
					// Fallback to config provided key if not resolved
					// Note: resolving happens outside usually, but here apiKey is passed resolved.
					// Ideally we resolve here again if needed?
					// The apiKey arg is resolved from semConfig.GetApiKey() which is deprecated.
					// We should resolve from openaiConf.GetApiKey() if present.
					// But executeSemantic resolves before calling factory.
					// Let's rely on caller or handle deprecated path.
				}
				return NewOpenAIEmbeddingProvider(key, openaiConf.GetModel()), nil
			}
			if conf.GetOllama() != nil {
				ollamaConf := conf.GetOllama()
				return NewOllamaEmbeddingProvider(ollamaConf.GetBaseUrl(), ollamaConf.GetModel()), nil
			}
			if conf.GetHttp() != nil {
				httpConf := conf.GetHttp()
				return NewHttpEmbeddingProvider(httpConf.GetUrl(), httpConf.GetHeaders(), httpConf.GetBodyTemplate(), httpConf.GetResponseJsonPath())
			}

			// Legacy/Deprecated path
			providerType := conf.GetProvider()
			model := conf.GetModel()

			if providerType == "openai" {
				return NewOpenAIEmbeddingProvider(apiKey, model), nil
			}
			return nil, fmt.Errorf("unknown provider: %s", providerType)
		},
	}
}

// SetProviderFactory allows overriding the default provider factory for testing.
func (m *CachingMiddleware) SetProviderFactory(factory ProviderFactory) {
	m.providerFactory = factory
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
		return m.executeSemantic(ctx, req, next, t, cacheConfig, labels)
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

func (m *CachingMiddleware) executeSemantic(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc, t tool.Tool, config *configv1.CacheConfig, labels []metrics.Label) (any, error) {
	serviceID := t.Tool().GetServiceId()
	semConfig := config.GetSemanticConfig()
	if semConfig == nil {
		logging.GetLogger().Warn("Semantic cache strategy selected but no semantic config provided", "tool", t.Tool().GetName())
		return next(ctx, req)
	}

	val, ok := m.semanticCaches.Load(serviceID)
	if !ok {
		// Resolve API Key
		// Priority:
		// 1. OpenAI Config Secret
		// 2. Deprecated API Key Secret
		var secret *configv1.SecretValue
		if semConfig.GetOpenai() != nil {
			secret = semConfig.GetOpenai().GetApiKey()
		} else {
			secret = semConfig.GetApiKey()
		}

		apiKey, err := util.ResolveSecret(ctx, secret)
		if err != nil {
			logging.GetLogger().Error("Failed to resolve semantic cache API key", "error", err)
			return next(ctx, req)
		}

		// Use factory to create provider
		provider, err := m.providerFactory(semConfig, apiKey)
		if err != nil {
			logging.GetLogger().Warn("Failed to create embedding provider", "error", err)
			return next(ctx, req)
		}

		newCache := NewSemanticCache(provider, semConfig.GetSimilarityThreshold())
		val, _ = m.semanticCaches.LoadOrStore(serviceID, newCache)
	}

	semCache := val.(*SemanticCache)
	inputStr := string(req.ToolInputs)

	// Inject CacheControl if not present
	var cacheControl *tool.CacheControl
	if cc, ok := tool.GetCacheControl(ctx); ok {
		cacheControl = cc
	} else {
		cacheControl = &tool.CacheControl{Action: tool.ActionAllow}
		ctx = tool.NewContextWithCacheControl(ctx, cacheControl)
	}

	var cached any
	var hit bool
	var embedding []float32
	var err error

	// Check cache ONLY if action is not DeleteCache
	if cacheControl.Action != tool.ActionDeleteCache {
		cached, embedding, hit, err = semCache.Get(ctx, req.ToolName, inputStr)
		if err == nil && hit {
			metrics.IncrCounterWithLabels([]string{"cache", "hits"}, 1, labels)
			return cached, nil
		} else if err != nil {
			logging.GetLogger().Error("Semantic cache error", "error", err)
		}
		if !hit {
			metrics.IncrCounterWithLabels([]string{"cache", "misses"}, 1, labels)
		}
	} else {
		metrics.IncrCounterWithLabels([]string{"cache", "skips"}, 1, labels)
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	if cacheControl.Action == tool.ActionDeleteCache {
		// Deletion not supported for semantic cache yet
		return result, nil
	}

	// Set cache if we have embedding
	if embedding != nil {
		if err := semCache.Set(ctx, req.ToolName, embedding, result, config.GetTtl().AsDuration()); err != nil {
			logging.GetLogger().Error("Failed to set semantic cache", "error", err)
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
	return m.cache.Clear(ctx)
}
