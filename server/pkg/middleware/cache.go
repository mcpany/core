// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

var cacheRedisClientCreator = redis.NewClient

// SetCacheRedisClientCreatorForTests sets the redis client creator for tests.
func SetCacheRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	cacheRedisClientCreator = creator
}

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	defaultCache *cache.Cache[any]
	toolManager  tool.ManagerInterface

	// caches stores initialized caches keyed by config hash
	// We use patrickmn/go-cache for storing *cache.Cache[any] instances
	caches *go_cache.Cache
	// redisClients stores initialized redis clients keyed by config hash
	redisClients sync.Map
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		defaultCache: cacheManager,
		toolManager:  toolManager,
		caches:       go_cache.New(1*time.Hour, 10*time.Minute),
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

	// Determine which cache to use
	cacheManager, err := m.getCache(ctx, cacheConfig)
	if err != nil {
		logging.GetLogger().Error("Failed to get cache manager", "error", err, "tool", t.Tool().GetName())
		return next(ctx, req)
	}

	// Extract service ID and Tool Name for metrics
	serviceID := t.Tool().GetServiceId()
	toolName := t.Tool().GetName()
	storageType := "memory"
	if cacheConfig.GetStorage() == configv1.CacheConfig_STORAGE_REDIS {
		storageType = "redis"
	}

	labels := []metrics.Label{
		{Name: "service", Value: serviceID},
		{Name: "tool", Value: toolName},
		{Name: "storage", Value: storageType},
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

	ttl := 5 * time.Minute // Default
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

func (m *CachingMiddleware) getCacheKey(req *tool.ExecutionRequest) string {
	return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
}

// Clear clears the default cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	// For now clear default.
	return m.defaultCache.Clear(ctx)
}

// Close closes all managed Redis clients.
func (m *CachingMiddleware) Close() error {
	var errs []error
	m.redisClients.Range(func(key, value any) bool {
		if client, ok := value.(*redis.Client); ok {
			if err := client.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		return true
	})
	if len(errs) > 0 {
		return fmt.Errorf("failed to close some redis clients: %v", errs)
	}
	return nil
}

func (m *CachingMiddleware) getCache(ctx context.Context, config *configv1.CacheConfig) (*cache.Cache[any], error) {
	if config.GetStorage() == configv1.CacheConfig_STORAGE_MEMORY || config.GetStorage() == configv1.CacheConfig_STORAGE_UNSPECIFIED {
		return m.defaultCache, nil
	}

	if config.GetStorage() == configv1.CacheConfig_STORAGE_REDIS {
		redisConfig := config.GetRedis()
		if redisConfig == nil {
			return nil, fmt.Errorf("redis config missing")
		}

		// Compute hash of redis config to reuse client
		configHash := hashRedisConfig(redisConfig)

		// Check if cache already exists
		if val, found := m.caches.Get(configHash); found {
			return val.(*cache.Cache[any]), nil
		}

		// Get or create Redis client
		client, err := m.getRedisClient(redisConfig, configHash)
		if err != nil {
			return nil, err
		}

		// Create Redis store
		redisStore := redis_store.NewRedis(client)
		cacheManager := cache.New[any](redisStore)

		// Cache the cache manager
		m.caches.Set(configHash, cacheManager, go_cache.DefaultExpiration)
		return cacheManager, nil
	}

	return m.defaultCache, nil
}

func hashRedisConfig(c *bus.RedisBus) string {
	s := fmt.Sprintf("%s|%d|%s", c.GetAddress(), c.GetDb(), c.GetPassword())
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (m *CachingMiddleware) getRedisClient(config *bus.RedisBus, hash string) (*redis.Client, error) {
	if val, ok := m.redisClients.Load(hash); ok {
		return val.(*redis.Client), nil
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := cacheRedisClientCreator(opts)
	m.redisClients.Store(hash, client)
	return client, nil
}
