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
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	go_cache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

// CachingMiddleware is a tool execution middleware that provides caching
// functionality.
type CachingMiddleware struct {
	// caches stores cache instances per service ID.
	caches       sync.Map // map[string]*cache.Cache[any]
	toolManager  tool.ManagerInterface
	defaultCache *cache.Cache[any]
}

// NewCachingMiddleware creates a new CachingMiddleware.
func NewCachingMiddleware(toolManager tool.ManagerInterface) *CachingMiddleware {
	goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
	cacheManager := cache.New[any](goCacheStore)
	return &CachingMiddleware{
		toolManager:  toolManager,
		defaultCache: cacheManager,
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

	// Get cache instance for this service
	c, err := m.getCache(t.Tool().GetServiceId(), cacheConfig)
	if err != nil {
		logging.GetLogger().Error("Failed to get cache instance", "error", err, "service", t.Tool().GetServiceId())
		// Fallback to no caching if cache creation fails
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

	// Check cache ONLY if action is not DeleteCache
	if cacheControl.Action != tool.ActionDeleteCache {
		// If normal allow (0), check cache.
		if cached, err := c.Get(ctx, cacheKey); err == nil {
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
		if err := c.Delete(ctx, cacheKey); err != nil {
			metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
			logging.GetLogger().Error("Failed to delete cache", "error", err, "tool", toolName)
		}
		return result, nil
	}

	if err := c.Set(ctx, cacheKey, result, store.WithExpiration(cacheConfig.GetTtl().AsDuration())); err != nil {
		metrics.IncrCounterWithLabels([]string{"cache", "errors"}, 1, labels)
		logging.GetLogger().Error("Failed to set cache", "error", err, "tool", toolName)
	}
	return result, nil
}

func (m *CachingMiddleware) getCache(serviceID string, config *configv1.CacheConfig) (*cache.Cache[any], error) {
	// Simple caching by ServiceID.
	// NOTE: If config changes dynamically, this cache might be stale.
	// ideally we should hash the config or handle reload.
	if val, ok := m.caches.Load(serviceID); ok {
		return val.(*cache.Cache[any]), nil
	}

	var cacheStore store.StoreInterface
	var clientToClose *redis.Client

	switch config.GetStorage() {
	case configv1.CacheConfig_STORAGE_REDIS:
		if config.GetRedis() == nil {
			return nil, fmt.Errorf("redis config missing")
		}
		opts := &redis.Options{
			Addr:     config.GetRedis().GetAddress(),
			Password: config.GetRedis().GetPassword(),
			DB:       int(config.GetRedis().GetDb()),
		}
		client := redisClientCreator(opts)
		clientToClose = client
		cacheStore = redis_store.NewRedis(client)
	case configv1.CacheConfig_STORAGE_MEMORY, configv1.CacheConfig_STORAGE_UNSPECIFIED:
		fallthrough
	default:
		// Default to memory
		goCacheStore := gocache_store.NewGoCache(go_cache.New(5*time.Minute, 10*time.Minute))
		cacheStore = goCacheStore
	}

	cacheManager := cache.New[any](cacheStore)

	// Use LoadOrStore to handle race condition
	actual, loaded := m.caches.LoadOrStore(serviceID, cacheManager)
	if loaded {
		// Race lost, use existing. Close the one we created if it has resources.
		if clientToClose != nil {
			_ = clientToClose.Close()
		}
		return actual.(*cache.Cache[any]), nil
	}

	return cacheManager, nil
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
	var input any
	if req.Arguments != nil {
		input = req.Arguments
	} else if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &input); err != nil {
			logging.GetLogger().Warn("Failed to unmarshal tool inputs for canonicalization", "error", err)
			return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
		}
	}

	inputs, err := util.CanonicalJSON(input)
	if err != nil {
		// Fallback if marshal fails (unlikely)
		logging.GetLogger().Warn("Failed to canonicalize tool inputs", "error", err)
		return fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
	}
	return fmt.Sprintf("%s:%s", req.ToolName, inputs)
}

// Clear clears the cache.
func (m *CachingMiddleware) Clear(ctx context.Context) error {
	// Clear default cache
	if err := m.defaultCache.Clear(ctx); err != nil {
		return err
	}
	// Clear all service caches
	var err error
	m.caches.Range(func(key, value any) bool {
		c := value.(*cache.Cache[any])
		if e := c.Clear(ctx); e != nil {
			err = e
			return false
		}
		return true
	})
	return err
}
