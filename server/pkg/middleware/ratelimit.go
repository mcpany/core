// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Limiter interface defines the methods required for a rate limiter.
type Limiter interface {
	// Allow checks if the request is allowed.
	Allow(ctx context.Context) (bool, error)
	// AllowN checks if the request is allowed with a specific cost.
	AllowN(ctx context.Context, n int) (bool, error)
	// Update updates the limiter configuration.
	Update(rps float64, burst int)
}

// LocalLimiter is an in-memory implementation of Limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
func (l *LocalLimiter) Update(rps float64, burst int) {
	limit := rate.Limit(rps)
	if l.Limit() != limit {
		l.SetLimit(limit)
	}
	if l.Burst() != burst {
		l.SetBurst(burst)
	}
}

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality for upstream services.
type RateLimitMiddleware struct {
	toolManager tool.ManagerInterface
	// limiters caches active limiters. Key is "serviceID:partitionKey".
	limiters *cache.Cache
	// redisClients caches Redis clients per service. Key is serviceID.
	redisClients sync.Map
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		toolManager: toolManager,
		limiters:    cache.New(1*time.Hour, 10*time.Minute),
	}
}

// Execute executes the rate limiting middleware.
func (m *RateLimitMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := m.toolManager.GetTool(req.ToolName)
	if !ok {
		return next(ctx, req)
	}

	serviceID := t.Tool().GetServiceId()
	serviceInfo, ok := m.toolManager.GetServiceInfo(serviceID)
	if !ok {
		// If service info is not found, we cannot enforce rate limits.
		// Proceed with execution.
		return next(ctx, req)
	}

	rateLimitConfig := serviceInfo.Config.GetRateLimit()
	if rateLimitConfig == nil || !rateLimitConfig.GetIsEnabled() {
		return next(ctx, req)
	}

	limiter, err := m.getLimiter(ctx, serviceID, rateLimitConfig)
	if err != nil {
		// Failed to get limiter (e.g. redis config error).
		// Fail open.
		return next(ctx, req)
	}

	// Calculate cost
	cost := 1
	if rateLimitConfig.GetCostMetric() == configv1.RateLimitConfig_COST_METRIC_TOKENS {
		cost = m.estimateTokenCost(req)
	}

	allowed, err := limiter.AllowN(ctx, cost)
	if err != nil {
		// If check fails (e.g. redis connection error).
		// Fail open.
		return next(ctx, req)
	}

	if !allowed {
		metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
			{Name: "service_id", Value: serviceID},
			{Name: "status", Value: "blocked"},
		})
		return nil, fmt.Errorf("rate limit exceeded for service %s", serviceInfo.Name)
	}

	metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
		{Name: "service_id", Value: serviceID},
		{Name: "status", Value: "allowed"},
	})

	return next(ctx, req)
}

func (m *RateLimitMiddleware) estimateTokenCost(req *tool.ExecutionRequest) int {
	// Use Arguments map if available, otherwise try to unmarshal ToolInputs
	var args map[string]interface{}

	if req.Arguments != nil {
		args = req.Arguments
	} else if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &args); err != nil {
			// If unmarshal fails, we can't estimate properly. Return default cost.
			return 1
		}
	} else {
		return 1
	}

	// Crude estimation: 1 token ~= 4 chars
	// Sum up all string arguments
	charCount := 0
	for _, v := range args {
		charCount += countChars(v)
	}

	// At least 1 token
	tokens := charCount / 4
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}

func countChars(v interface{}) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []interface{}:
		count := 0
		for _, item := range val {
			count += countChars(item)
		}
		return count
	case map[string]interface{}:
		count := 0
		for _, item := range val {
			count += countChars(item)
		}
		return count
	default:
		// Best effort for other types
		return len(fmt.Sprintf("%v", val))
	}
}

func (m *RateLimitMiddleware) getLimiter(ctx context.Context, serviceID string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	partitionKey := m.getPartitionKey(ctx, config.GetKeyBy())
	cacheKey := serviceID
	if partitionKey != "" {
		cacheKey = fmt.Sprintf("%s:%s", serviceID, partitionKey)
	}

	isRedis := config.GetStorage() == configv1.RateLimitConfig_STORAGE_REDIS

	// Try to get from cache
	if val, found := m.limiters.Get(cacheKey); found {
		limiter := val.(Limiter)
		// Verify type matches config
		validType := false
		if isRedis {
			_, validType = limiter.(*RedisLimiter)
		} else {
			_, validType = limiter.(*LocalLimiter)
		}

		if validType {
			// Update config in case it changed
			limiter.Update(rps, burst)
			return limiter, nil
		}
		// Type mismatch, fall through to create new
	}

	// Create new limiter
	var limiter Limiter

	if isRedis {
		if config.GetRedis() == nil {
			return nil, fmt.Errorf("redis config is missing")
		}
		client, err := m.getRedisClient(serviceID, config.GetRedis())
		if err != nil {
			return nil, err
		}
		limiter = NewRedisLimiterWithClient(client, serviceID, partitionKey, config)
	} else {
		limiter = &LocalLimiter{
			Limiter: rate.NewLimiter(rate.Limit(rps), burst),
		}
	}

	// Cache it
	m.limiters.Set(cacheKey, limiter, cache.DefaultExpiration)

	return limiter, nil
}

func (m *RateLimitMiddleware) getPartitionKey(ctx context.Context, keyBy configv1.RateLimitConfig_KeyBy) string {
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
		// Fallback or separate bucket for anonymous?
		return "user:anonymous"
	case configv1.RateLimitConfig_KEY_BY_API_KEY:
		if apiKey, ok := auth.APIKeyFromContext(ctx); ok {
			return hashKey("apikey:", apiKey)
		}
		// Fallback to extraction from HTTP request if available
		if req, ok := ctx.Value("http.request").(*http.Request); ok {
			if key := req.Header.Get("X-API-Key"); key != "" {
				return hashKey("apikey:", key)
			}
			if key := req.Header.Get("Authorization"); key != "" {
				// Use hash of token to avoid storing sensitive data in cache keys
				return hashKey("auth:", key)
			}
		}
		return ""
	default:
		return ""
	}
}

func hashKey(prefix, key string) string {
	hash := sha256.Sum256([]byte(key))
	// Pre-allocate buffer for prefix + hex encoded hash (32 bytes -> 64 hex chars)
	buf := make([]byte, len(prefix)+hex.EncodedLen(len(hash)))
	copy(buf, prefix)
	hex.Encode(buf[len(prefix):], hash[:])
	return string(buf)
}

func (m *RateLimitMiddleware) getRedisClient(serviceID string, config *bus.RedisBus) (*redis.Client, error) {
	if val, ok := m.redisClients.Load(serviceID); ok {
		return val.(*redis.Client), nil
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := redisClientCreator(opts)
	m.redisClients.Store(serviceID, client)
	return client, nil
}
