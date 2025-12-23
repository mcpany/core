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
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	// caches maps serviceID to *cacheEntry
	caches      *go_cache.Cache
	toolManager tool.ManagerInterface
	// mu protects creation of new cache managers to avoid race conditions
	// However, go-cache is thread safe. But checking and setting is not atomic.
	// We can use a map of mutexes or a single global mutex for creation.
	// Given creation is rare (once per service), a global mutex or sync.Map + mutex/once is fine.
	// We use a keyed mutex approach implicitly via singleflight or just a mutex if we want to be simple.
	// But `go_cache` handles expiration.
	// Let's use `sync.Map` for locks?
	// Or just a global lock for `getCacheManager` if we accept slightly more contention on MISS?
	// It's cleaner to lock only when creating.
	creationMu sync.Mutex
}

type cacheEntry struct {
	manager     *cache.Cache[any]
	redisClient *redis.Client // Keep track to close it
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	m := &CachingMiddleware{
		caches:      go_cache.New(1*time.Hour, 10*time.Minute),
		toolManager: toolManager,
	}

	// Register an eviction callback to close Redis clients
	m.caches.OnEvicted(func(s string, i interface{}) {
		if entry, ok := i.(*cacheEntry); ok {
			if entry.redisClient != nil {
				logging.GetLogger().Info("Closing Redis client for evicted service cache", "service", s)
				if err := entry.redisClient.Close(); err != nil {
					logging.GetLogger().Error("Failed to close Redis client", "error", err, "service", s)
				}
			}
		}
	})

	return m
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

	serviceID := t.Tool().GetServiceId()
	toolName := t.Tool().GetName()
	labels := []metrics.Label{
		{Name: "service", Value: serviceID},
		{Name: "tool", Value: toolName},
	}

	// Get or create cache manager for this service
	cacheManager, err := m.getCacheManager(serviceID, cacheConfig)
	if err != nil {
		logging.GetLogger().Error("Failed to get cache manager", "error", err, "service", serviceID)
		// Fail open: execute without cache
		return next(ctx, req)
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
		if cached, err := cacheManager.Get(ctx, cacheKey); err == nil {
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
		if err := cacheManager.Delete(ctx, cacheKey); err != nil {
			metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
			logging.GetLogger().Error("Failed to delete cache", "error", err, "tool", toolName)
		}
		return result, nil
	}

	// Default TTL from config or fallback
	ttl := 5 * time.Minute
	if cacheConfig.GetTtl() != nil {
		ttl = cacheConfig.GetTtl().AsDuration()
	}

	if err := cacheManager.Set(ctx, cacheKey, result, store.WithExpiration(ttl)); err != nil {
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

func (m *CachingMiddleware) getCacheManager(serviceID string, config *configv1.CacheConfig) (*cache.Cache[any], error) {
	// First check without lock
	if val, found := m.caches.Get(serviceID); found {
		return val.(*cacheEntry).manager, nil
	}

	// Lock to ensure single creation
	m.creationMu.Lock()
	defer m.creationMu.Unlock()

	// Double check
	if val, found := m.caches.Get(serviceID); found {
		return val.(*cacheEntry).manager, nil
	}

	// Create new manager
	var cacheStore store.StoreInterface
	var redisClient *redis.Client

	switch config.GetStorage() {
	case configv1.CacheConfig_STORAGE_REDIS:
		if config.GetRedis() == nil {
			return nil, fmt.Errorf("redis config missing")
		}
		redisOpts := &redis.Options{
			Addr:     config.GetRedis().GetAddress(),
			Password: config.GetRedis().GetPassword(),
			DB:       int(config.GetRedis().GetDb()),
		}
		// Reuse redisClientCreator to allow mocking in tests
		redisClient = redisClientCreator(redisOpts)
		cacheStore = redis_store.NewRedis(redisClient)

	case configv1.CacheConfig_STORAGE_MEMORY, configv1.CacheConfig_STORAGE_UNSPECIFIED:
		fallthrough
	default:
		cacheStore = gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	}

	cacheManager := cache.New[any](cacheStore)
	m.caches.Set(serviceID, &cacheEntry{
		manager:     cacheManager,
		redisClient: redisClient,
	}, go_cache.DefaultExpiration)

	return cacheManager, nil
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

// Clear clears the cache (this clears the map of managers, not necessarily the underlying stores)
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	m.caches.Flush()
	return nil
}
