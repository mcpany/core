// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	go_cache "github.com/patrickmn/go-cache"
)

// CachingSampler wraps a tool.Sampler to provide caching capabilities.
type CachingSampler struct {
	sampler tool.Sampler
	cache   *cache.Cache[any]
	config  *configv1.CacheConfig
}

// NewCachingSampler creates a new CachingSampler.
func NewCachingSampler(sampler tool.Sampler, config *configv1.CacheConfig) *CachingSampler {
	ttl := 5 * time.Minute
	if config != nil && config.GetTtl().AsDuration() > 0 {
		ttl = config.GetTtl().AsDuration()
	}

	goCacheStore := gocache_store.NewGoCache(go_cache.New(ttl, ttl*2))
	cacheManager := cache.New[any](goCacheStore)

	return &CachingSampler{
		sampler: sampler,
		cache:   cacheManager,
		config:  config,
	}
}

// CreateMessage requests a message creation from the client (sampling), checking cache first.
func (s *CachingSampler) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.config == nil || !s.config.GetIsEnabled() {
		return s.sampler.CreateMessage(ctx, params)
	}

	// Skip caching if context is included, as the response depends on client state not captured in params.
	if params.IncludeContext != "" && params.IncludeContext != "none" {
		logging.GetLogger().Debug("Skipping sampling cache because context is included", "includeContext", params.IncludeContext)
		return s.sampler.CreateMessage(ctx, params)
	}

	cacheKey, err := s.getCacheKey(params)
	if err != nil {
		logging.GetLogger().Warn("Failed to generate cache key for sampling", "error", err)
		return s.sampler.CreateMessage(ctx, params)
	}

	// Try to get from cache
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if result, ok := cached.(*mcp.CreateMessageResult); ok {
			logging.GetLogger().Debug("Sampling cache hit", "key", cacheKey)
			return result, nil
		}
	}

	// Cache miss
	result, err := s.sampler.CreateMessage(ctx, params)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ttl := 5 * time.Minute
	if s.config.GetTtl().AsDuration() > 0 {
		ttl = s.config.GetTtl().AsDuration()
	}

	if err := s.cache.Set(ctx, cacheKey, result, store.WithExpiration(ttl)); err != nil {
		logging.GetLogger().Warn("Failed to set sampling cache", "error", err)
	}

	return result, nil
}

func (s *CachingSampler) getCacheKey(params *mcp.CreateMessageParams) (string, error) {
	// We hash the JSON representation of the params
	// Note: We should probably normalize the params (sort keys, etc) but json.Marshal
	// does map key sorting.
	// We must ensure that non-deterministic fields are handled if any.
	// CreateMessageParams contains Messages, ModelPreferences, SystemPrompt, etc.
	// All should be part of the key.

	// However, `IncludeContext` might contain large data.
	// If `params.IncludeContext` is "allServers" or "thisServer", it just affects what context is sent.
	// It doesn't change the request per se, but the *response* depends on the context provided by the CLIENT.
	// Wait, Sampling is Server -> Client.
	// The Server sends `CreateMessageParams`. The Client responds.
	// The `params` define what the Server is asking.
	// If `IncludeContext` is requested, the Client will include context in its processing.
	// So `IncludeContext` should be part of the key.

	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return "sampling:" + hex.EncodeToString(hash[:]), nil
}

// Verify that CachingSampler implements tool.Sampler
var _ tool.Sampler = (*CachingSampler)(nil)
