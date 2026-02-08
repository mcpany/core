// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	jsoniter "github.com/json-iterator/go"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
)

// ProviderFactory is a function that creates an EmbeddingProvider.
//
// Summary: is a function that creates an EmbeddingProvider.
type ProviderFactory func(config *configv1.SemanticCacheConfig, apiKey string) (EmbeddingProvider, error)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
var (
	metricCacheHits   = []string{"cache", "hits"}
	metricCacheMisses = []string{"cache", "misses"}
	metricCacheSkips  = []string{"cache", "skips"}
	metricCacheErrors = []string{"cache", "errors"}
)

// CachingMiddleware handles caching of tool execution results.
//
// Summary: handles caching of tool execution results.
type CachingMiddleware struct {
	cache           *cache.Cache[any]
	toolManager     tool.ManagerInterface
	semanticCaches  sync.Map
	initMu          sync.Mutex // Guards semantic cache initialization
	providerFactory ProviderFactory
	hasherPool      *sync.Pool
}

// NewCachingMiddleware creates a new CachingMiddleware.
//
// Summary: creates a new CachingMiddleware.
//
// Parameters:
//   - toolManager: tool.ManagerInterface. The toolManager.
//
// Returns:
//   - *CachingMiddleware: The *CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		cache:       cacheManager,
		toolManager: toolManager,
		hasherPool: &sync.Pool{
			New: func() any {
				return fnv.New128a()
			},
		},
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
					_ = "noop" // Satisfy SA9003
				}
				return NewOpenAIEmbeddingProvider(key, openaiConf.GetModel()), nil
			}
			if conf.GetOllama() != nil {
				ollamaConf := conf.GetOllama()
				return NewOllamaEmbeddingProvider(ollamaConf.GetBaseUrl(), ollamaConf.GetModel()), nil
			}
			if conf.GetHttp() != nil {
				httpConf := conf.GetHttp()
				provider, err := NewHTTPEmbeddingProvider(
					httpConf.GetUrl(),
					httpConf.GetHeaders(),
					httpConf.GetBodyTemplate(),
					httpConf.GetResponseJsonPath(),
				)
				if err != nil {
					return nil, err
				}
				return provider, nil
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
//
// Summary: allows overriding the default provider factory for testing.
//
// Parameters:
//   - factory: ProviderFactory. The factory.
//
// Returns:
//   None.
func (m *CachingMiddleware) SetProviderFactory(factory ProviderFactory) {
	m.providerFactory = factory
}

// Execute executes the caching middleware.
//
// Summary: executes the caching middleware.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *tool.ExecutionRequest. The req.
//   - next: tool.ExecutionFunc. The next.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
			metrics.IncrCounterWithLabels(metricCacheHits, 1, labels)
			return cached, nil
		}
		// Not found in cache
		metrics.IncrCounterWithLabels(metricCacheMisses, 1, labels)
	} else {
		// If DeleteCache, we skip cache lookup (force miss)
		metrics.IncrCounterWithLabels(metricCacheSkips, 1, labels)
	}

	result, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check CacheControl
	if cacheControl.Action == tool.ActionDeleteCache {
		if err := m.cache.Delete(ctx, cacheKey); err != nil {
			metrics.IncrCounterWithLabels(metricCacheErrors, 1, labels)
			logging.GetLogger().Error("Failed to delete cache", "error", err, "tool", toolName)
		}
		return result, nil
	}

	if err := m.cache.Set(ctx, cacheKey, result, store.WithExpiration(cacheConfig.GetTtl().AsDuration())); err != nil {
		metrics.IncrCounterWithLabels(metricCacheErrors, 1, labels)
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
		// Double-checked locking to prevent race condition and connection leaks
		m.initMu.Lock()
		val, ok = m.semanticCaches.Load(serviceID)
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
				m.initMu.Unlock()
				logging.GetLogger().Error("Failed to resolve semantic cache API key", "error", err)
				return next(ctx, req)
			}

			// Use factory to create provider
			provider, err := m.providerFactory(semConfig, apiKey)
			if err != nil {
				m.initMu.Unlock()
				logging.GetLogger().Warn("Failed to create embedding provider", "error", err)
				return next(ctx, req)
			}

			var vectorStore VectorStore
			persistencePath := semConfig.GetPersistencePath()
			if persistencePath != "" {
				var err error
				if strings.HasPrefix(persistencePath, "postgres://") || strings.HasPrefix(persistencePath, "postgresql://") {
					vectorStore, err = NewPostgresVectorStore(persistencePath)
					if err != nil {
						logging.GetLogger().Error("Failed to create Postgres vector store", "error", err, "dsn", persistencePath)
						// Fallback to in-memory
						vectorStore = NewSimpleVectorStore()
					}
				} else {
					vectorStore, err = NewSQLiteVectorStore(persistencePath)
					if err != nil {
						logging.GetLogger().Error("Failed to create SQLite vector store", "error", err, "path", persistencePath)
						// Fallback to in-memory
						vectorStore = NewSimpleVectorStore()
					}
				}
			} else {
				vectorStore = NewSimpleVectorStore()
			}

			newCache := NewSemanticCache(provider, vectorStore, semConfig.GetSimilarityThreshold())
			m.semanticCaches.Store(serviceID, newCache)
			val = newCache
		}
		m.initMu.Unlock()
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
			metrics.IncrCounterWithLabels(metricCacheHits, 1, labels)
			return cached, nil
		} else if err != nil {
			logging.GetLogger().Error("Semantic cache error", "error", err)
		}
		if !hit {
			metrics.IncrCounterWithLabels(metricCacheMisses, 1, labels)
		}
	} else {
		metrics.IncrCounterWithLabels(metricCacheSkips, 1, labels)
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
		// ⚡ BOLT: Offload cache write to background to avoid blocking the response.
		// Randomized Selection from Top 5 High-Impact Targets
		go func() {
			// Detach context so write completes even if request context is canceled.
			// Add timeout to prevent hanging.
			bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()

			if err := semCache.Set(bgCtx, req.ToolName, embedding, result, config.GetTtl().AsDuration()); err != nil {
				logging.GetLogger().Error("Failed to set semantic cache", "error", err)
			}
		}()
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
	var normalizedInputs []byte
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	// Optimization: Use Arguments map if available to avoid unnecessary unmarshal/marshal cycle.
	// json.Marshal sorts map keys, so it produces a canonical representation for caching.
	if req.Arguments != nil {
		if marshaled, err := json.Marshal(req.Arguments); err == nil {
			normalizedInputs = marshaled
		}
	}

	// Normalize ToolInputs if they are JSON and we couldn't use Arguments
	if normalizedInputs == nil && len(req.ToolInputs) > 0 {
		// Optimization: Check if it looks like a JSON object or array before unmarshaling
		// Skip leading whitespace (simplified check)
		var firstChar byte
		for _, b := range req.ToolInputs {
			if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
				firstChar = b
				break
			}
		}

		if firstChar == '{' || firstChar == '[' {
			var input any
			// We use standard json.Unmarshal which sorts map keys when Marshaling back.
			if err := json.Unmarshal(req.ToolInputs, &input); err == nil {
				if marshaled, err := json.Marshal(input); err == nil {
					normalizedInputs = marshaled
				}
			}
		}
	}

	// Fallback to raw bytes if unmarshal fails or empty
	if normalizedInputs == nil {
		normalizedInputs = req.ToolInputs
	}

	// Optimization: Hash the normalized inputs to keep the cache key short and fixed length.
	// This avoids using potentially large JSON strings as map keys.
	// We use FNV-1a 128-bit hash which is significantly faster than SHA256 (>2x)
	// and produces a shorter key (32 hex chars vs 64), saving memory.
	// We use a sync.Pool to reuse hashers and avoid heap allocations.
	h := m.hasherPool.Get().(hash.Hash)
	defer m.hasherPool.Put(h)
	h.Reset()
	_, _ = h.Write(normalizedInputs)

	// FNV-1a 128-bit hash is 16 bytes. Use stack buffer to avoid allocation.
	var hashBuf [16]byte
	hashBytes := h.Sum(hashBuf[:0])

	// Optimization: Use strings.Builder to build the key to avoid one allocation (string conversion)
	// and use stack buffer for hex encoding.
	var sb strings.Builder
	// 32 is hex encoded length of 16 bytes
	sb.Grow(len(req.ToolName) + 1 + 32)
	sb.WriteString(req.ToolName)
	sb.WriteByte(':')

	var hexBuf [32]byte
	encodedLen := hex.Encode(hexBuf[:], hashBytes)
	sb.Write(hexBuf[:encodedLen])

	return sb.String()
}

// Clear clears the cache.
//
// Summary: clears the cache.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	return m.cache.Clear(ctx)
}
