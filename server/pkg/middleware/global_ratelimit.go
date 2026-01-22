// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// GlobalRateLimitMiddleware provides rate limiting functionality for all MCP requests.
type GlobalRateLimitMiddleware struct {
	mu     sync.RWMutex
	config *configv1.RateLimitConfig
	// limiters caches active limiters. Key is "partitionKey".
	limiters *cache.Cache
	// redisClients caches Redis clients. Key is "global".
	redisClients sync.Map
}

// NewGlobalRateLimitMiddleware creates a new GlobalRateLimitMiddleware.
//
// config holds the configuration settings.
//
// Returns the result.
func NewGlobalRateLimitMiddleware(config *configv1.RateLimitConfig) *GlobalRateLimitMiddleware {
	return &GlobalRateLimitMiddleware{
		config:   config,
		limiters: cache.New(1*time.Hour, 10*time.Minute),
	}
}

// UpdateConfig updates the rate limit configuration safely.
//
// config holds the configuration settings.
func (m *GlobalRateLimitMiddleware) UpdateConfig(config *configv1.RateLimitConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	// We might want to clear limiters cache if config changes drastically,
	// but limiters update themselves on access if rps/burst changes.
	// So we generally don't need to clear cache unless KeyBy changes.
}

// Execute executes the rate limiting middleware.
//
// ctx is the context for the request.
// method is the method.
// req is the request object.
// next is the next.
//
// Returns the result.
// Returns an error if the operation fails.
func (m *GlobalRateLimitMiddleware) Execute(ctx context.Context, method string, req mcp.Request, next mcp.MethodHandler) (mcp.Result, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()

	if config == nil || !config.GetIsEnabled() {
		return next(ctx, method, req)
	}

	limiter, err := m.getLimiter(ctx, config)
	if err == nil {
		allowed, err := limiter.Allow(ctx)
		if err != nil {
			// Fail open on error
			return next(ctx, method, req)
		}
		if !allowed {
			m.recordMetrics("blocked")
			return nil, fmt.Errorf("global rate limit exceeded")
		}
		m.recordMetrics("allowed")
	}

	return next(ctx, method, req)
}

func (m *GlobalRateLimitMiddleware) recordMetrics(status string) {
	metrics.IncrCounterWithLabels([]string{"global_rate_limit", "requests_total"}, 1, []metrics.Label{
		{Name: "status", Value: status},
	})
}

// getLimiter retrieves or creates a limiter.
func (m *GlobalRateLimitMiddleware) getLimiter(ctx context.Context, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	partitionKey := m.getPartitionKey(ctx, config.GetKeyBy())

	// Cache key: partitionKey
	cacheKey := partitionKey

	isRedis := config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS

	// Try to get from cache
	if val, found := m.limiters.Get(cacheKey); found {
		limiter := val.(Limiter)
		// Verify type matches config
		var validType bool
		if isRedis {
			rl, ok := limiter.(*RedisLimiter)
			validType = ok
			// Check if Redis config changed
			if ok && config.GetRedis() != nil {
				configHash := m.calculateConfigHash(config.GetRedis())
				if rl.GetConfigHash() != configHash {
					validType = false // Force creation of new limiter
				}
			}
		} else {
			_, validType = limiter.(*LocalLimiter)
		}

		if validType {
			// Update config in case it changed
			limiter.Update(rps, burst)
			return limiter, nil
		}
		// Type mismatch or config changed, fall through to create new
	}

	// Create new limiter
	var limiter Limiter

	if isRedis {
		if config.GetRedis() == nil {
			return nil, fmt.Errorf("redis config is missing")
		}
		client := m.getRedisClient(config.GetRedis())
		// Pass global identifier
		limiter = NewRedisLimiterWithClient(client, "global", "", partitionKey, config)
	} else {
		limiter = &LocalLimiter{
			Limiter: rate.NewLimiter(rate.Limit(rps), burst),
		}
	}

	// Cache it
	m.limiters.Set(cacheKey, limiter, cache.DefaultExpiration)

	return limiter, nil
}

func (m *GlobalRateLimitMiddleware) getPartitionKey(ctx context.Context, keyBy configv1.RateLimitConfig_KeyBy) string {
	switch keyBy {
	case configv1.RateLimitConfig_KEY_BY_IP:
		if ip, ok := util.RemoteIPFromContext(ctx); ok {
			return "ip:" + ip
		}
		return "ip:unknown"
	case configv1.RateLimitConfig_KEY_BY_USER_ID:
		if uid, ok := auth.UserFromContext(ctx); ok {
			return "user:" + uid
		}
		return "user:anonymous"
	case configv1.RateLimitConfig_KEY_BY_API_KEY:
		if apiKey, ok := auth.APIKeyFromContext(ctx); ok {
			return hashKey("apikey:", apiKey)
		}
		// Fallback to extraction from HTTP request if available
		if req, ok := ctx.Value(HTTPRequestContextKey).(*http.Request); ok {
			if key := req.Header.Get("X-API-Key"); key != "" {
				return hashKey("apikey:", key)
			}
			if key := req.Header.Get("Authorization"); key != "" {
				// Use hash of token to avoid storing sensitive data in cache keys
				return hashKey("auth:", key)
			}
		}
		return "apikey:none"
	case configv1.RateLimitConfig_KEY_BY_GLOBAL:
		return "global"
	default:
		// Default to global
		return "global"
	}
}

func (m *GlobalRateLimitMiddleware) calculateConfigHash(config *bus.RedisBus) string {
	// Hash the sensitive config to avoid storing passwords in memory as clear text keys if possible
	data := config.GetAddress() + "|" + config.GetPassword() + "|" + strconv.Itoa(int(config.GetDb()))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (m *GlobalRateLimitMiddleware) getRedisClient(config *bus.RedisBus) *redis.Client {
	configHash := m.calculateConfigHash(config)

	if val, ok := m.redisClients.Load("global"); ok {
		if cached, ok := val.(*cachedRedisClient); ok {
			if cached.configHash == configHash {
				return cached.client
			}
		}
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := redisClientCreator(opts)
	m.redisClients.Store("global", &cachedRedisClient{
		client:     client,
		configHash: configHash,
	})
	return client
}
