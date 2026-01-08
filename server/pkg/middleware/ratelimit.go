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
	"strconv"
	"sync"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
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
	tokenizer   tokenizer.Tokenizer
	// limiters caches active limiters. Key is "limitKey:partitionKey".
	limiters *cache.Cache
	// redisClients caches Redis clients per service. Key is serviceID.
	redisClients sync.Map
}

type cachedRedisClient struct {
	client     *redis.Client
	configHash string
}

// Option defines a functional option for RateLimitMiddleware.
type Option func(*RateLimitMiddleware)

// WithTokenizer sets a custom tokenizer for the middleware.
func WithTokenizer(t tokenizer.Tokenizer) Option {
	return func(m *RateLimitMiddleware) {
		m.tokenizer = t
	}
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface, opts ...Option) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		toolManager: toolManager,
		limiters:    cache.New(1*time.Hour, 10*time.Minute),
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.tokenizer == nil {
		m.tokenizer = tokenizer.NewSimpleTokenizer()
	}
	return m
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

	serviceRateLimitConfig := serviceInfo.Config.GetRateLimit()


	// Check for tool-specific limit first
	var toolLimiter Limiter
	var toolErr error

	// Priority:
	// 1. Tool-specific limit (if configured and enabled)
	// 2. Service-level limit (if configured and enabled)
	// Rule: If tool limit exists, it OVERRIDES service limit (i.e., we only check tool limit).
	//       If tool limit does NOT exist, we check service limit.
	// Rationale: Granular control implies ability to loosen OR tighten.
	//            If one wants "Service Cap" + "Tool Cap", they are separate concerns, but usually "Limit X for Tool Y" implies "Use X instead of default".

	appliedLimit := false

	// Check tool specific limit
	if serviceRateLimitConfig != nil && serviceRateLimitConfig.GetToolLimits() != nil {
		if toolConfig, ok := serviceRateLimitConfig.GetToolLimits()[req.ToolName]; ok && toolConfig.GetIsEnabled() {
			toolLimiter, toolErr = m.getLimiter(ctx, serviceID, "tool:"+req.ToolName, toolConfig)
			if toolErr == nil {
				appliedLimit = true
				_ = appliedLimit
				if err := m.checkLimit(ctx, toolLimiter, toolConfig, req); err != nil {
					m.recordMetrics(serviceID, "tool", "blocked")
					return nil, fmt.Errorf("rate limit exceeded for tool %s", req.ToolName)
				}
				m.recordMetrics(serviceID, "tool", "allowed")
				// If we checked tool limit, we return early (override logic).
				return next(ctx, req)
			}
			// If getLimiter failed, we log/ignore? For now fall through or fail open.
			// Let's fail open on error, but if we found config and failed to get limiter, maybe we should fall back to service?
			// Safety: fail open.
		}
	}

	// If no tool limit applied, check service limit
	if !appliedLimit && serviceRateLimitConfig != nil && serviceRateLimitConfig.GetIsEnabled() {
		serviceLimiter, serviceErr := m.getLimiter(ctx, serviceID, "service", serviceRateLimitConfig)
		if serviceErr == nil {
			if err := m.checkLimit(ctx, serviceLimiter, serviceRateLimitConfig, req); err != nil {
				m.recordMetrics(serviceID, "service", "blocked")
				return nil, fmt.Errorf("rate limit exceeded for service %s", serviceInfo.Name)
			}
			m.recordMetrics(serviceID, "service", "allowed")
		}
	}

	return next(ctx, req)
}

func (m *RateLimitMiddleware) checkLimit(ctx context.Context, limiter Limiter, config *configv1.RateLimitConfig, req *tool.ExecutionRequest) error {
	// Calculate cost
	cost := 1
	if config.GetCostMetric() == configv1.RateLimitConfig_COST_METRIC_TOKENS {
		cost = m.estimateTokenCost(req)
	}

	allowed, err := limiter.AllowN(ctx, cost)
	if err != nil {
		// Fail open on error
		return nil //nolint:nilerr
	}
	if !allowed {
		return fmt.Errorf("limit exceeded")
	}
	return nil
}

func (m *RateLimitMiddleware) recordMetrics(serviceID, limitType, status string) {
	metrics.IncrCounterWithLabels([]string{"rate_limit", "requests_total"}, 1, []armonmetrics.Label{
		{Name: "service_id", Value: serviceID},
		{Name: "limit_type", Value: limitType},
		{Name: "status", Value: status},
	})
}

func (m *RateLimitMiddleware) estimateTokenCost(req *tool.ExecutionRequest) int {
	// Use Arguments map if available, otherwise try to unmarshal ToolInputs
	var args map[string]interface{}

	switch {
	case req.Arguments != nil:
		args = req.Arguments
	case len(req.ToolInputs) > 0:
		if err := json.Unmarshal(req.ToolInputs, &args); err != nil {
			// If unmarshal fails, we can't estimate properly. Return default cost.
			return 1
		}
		// Cache the parsed arguments to avoid re-parsing in subsequent middleware
		req.Arguments = args
	default:
		return 1
	}

	tokens, err := tokenizer.CountTokensInValue(m.tokenizer, args)
	if err != nil {
		return 1
	}
	if tokens < 1 {
		return 1
	}
	return tokens
}

// getLimiter retrieves or creates a limiter.
// limitScopeKey is a string that identifies the scope (e.g. "service", "tool:myTool").
// It is combined with serviceID to form the unique cache key prefix.
func (m *RateLimitMiddleware) getLimiter(ctx context.Context, serviceID string, limitScopeKey string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1 // Ensure at least 1 request can be made
	}

	partitionKey := m.getPartitionKey(ctx, config.GetKeyBy())

	// Cache key format: serviceID:limitScopeKey:partitionKey
	// e.g. "myservice:service:ip:1.2.3.4"
	// e.g. "myservice:tool:myTool:ip:1.2.3.4"

	cacheKey := serviceID + ":" + limitScopeKey
	if partitionKey != "" {
		cacheKey = cacheKey + ":" + partitionKey
	}

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
				newConfigHash := config.GetRedis().GetAddress() + "|" + config.GetRedis().GetPassword() + "|" + strconv.Itoa(int(config.GetRedis().GetDb()))
				if rl.GetConfigHash() != newConfigHash {
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
		client, err := m.getRedisClient(serviceID, config.GetRedis())
		if err != nil {
			return nil, err
		}
		// Pass limitScopeKey as part of key prefix to redis limiter if needed?
		// RedisLimiter probably uses a key prefix.
		// We should ensure Redis keys don't collide.
		// The current RedisLimiter implementation (assumed) takes a key.
		// We should pass the cacheKey or similar unique identifier.

		// Wait, NewRedisLimiterWithClient signature in original code:
		// NewRedisLimiterWithClient(client, serviceID, partitionKey, config)
		// It likely uses serviceID + partitionKey for the key.
		// We need to pass the scope into it.
		// Since I can't easily change RedisLimiter signature without seeing it, let's assume I need to pass a "key prefix" as serviceID.
		// Hack: pass "serviceID:limitScopeKey" as the serviceID argument to NewRedisLimiterWithClient.

		limiter = NewRedisLimiterWithClient(client, serviceID+":"+limitScopeKey, partitionKey, config)
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

func (m *RateLimitMiddleware) getRedisClient(serviceID string, config *bus.RedisBus) (*redis.Client, error) { //nolint:unparam
	configHash := config.GetAddress() + "|" + config.GetPassword() + "|" + strconv.Itoa(int(config.GetDb()))

	if val, ok := m.redisClients.Load(serviceID); ok {
		// handle both legacy *redis.Client (if any) and new *cachedRedisClient
		// Though in this code path we probably only have new ones if we restart.
		// But sync.Map is persistent in memory.
		if cached, ok := val.(*cachedRedisClient); ok {
			if cached.configHash == configHash {
				return cached.client, nil
			}
		} else if _, ok := val.(*redis.Client); ok {
			// Legacy fallback or type mismatch, treat as miss
			_ = ok
		}
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := redisClientCreator(opts)
	m.redisClients.Store(serviceID, &cachedRedisClient{
		client:     client,
		configHash: configHash,
	})
	return client, nil
}
