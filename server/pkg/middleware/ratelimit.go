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
	"time"

	armonmetrics "github.com/armon/go-metrics"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/patrickmn/go-cache"
)

// metricRateLimitRequestsTotal is the metric name for rate limit requests.
// Pre-allocated to avoid allocation on every request.
var metricRateLimitRequestsTotal = []string{"rate_limit", "requests_total"}

// RateLimitMiddleware is a tool execution middleware that provides rate limiting
// functionality for upstream services.
type RateLimitMiddleware struct {
	toolManager tool.ManagerInterface
	tokenizer   tokenizer.Tokenizer
	// limiters caches active limiters. Key is "limitKey:partitionKey".
	limiters *cache.Cache
	// strategies maps storage types to strategies.
	strategies map[configv1.RateLimitConfig_Storage]RateLimitStrategy
}

// Option defines a functional option for RateLimitMiddleware.
type Option func(*RateLimitMiddleware)

// WithTokenizer sets a custom tokenizer for the middleware.
//
// Summary: Functional option to set a custom tokenizer.
//
// Parameters:
//   - t: tokenizer.Tokenizer. The tokenizer instance.
//
// Returns:
//   - Option: The configured option.
func WithTokenizer(t tokenizer.Tokenizer) Option {
	return func(m *RateLimitMiddleware) {
		m.tokenizer = t
	}
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
//
// Summary: Creates a new RateLimitMiddleware instance.
//
// Parameters:
//   - toolManager: tool.ManagerInterface. The tool manager to retrieve service info.
//   - opts: ...Option. Optional configuration functions.
//
// Returns:
//   - *RateLimitMiddleware: The new RateLimitMiddleware instance.
func NewRateLimitMiddleware(toolManager tool.ManagerInterface, opts ...Option) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		toolManager: toolManager,
		limiters:    cache.New(1*time.Hour, 10*time.Minute),
		strategies:  make(map[configv1.RateLimitConfig_Storage]RateLimitStrategy),
	}

	// Register default strategies
	m.strategies[configv1.RateLimitConfig_STORAGE_MEMORY] = NewLocalStrategy()
	// Redis strategy requires a client provider or we can use the default one if it manages clients internally.
	// For now, let's assume we want to use the one that manages clients.
	m.strategies[configv1.RateLimitConfig_STORAGE_REDIS] = NewRedisStrategy()

	for _, opt := range opts {
		opt(m)
	}
	if m.tokenizer == nil {
		m.tokenizer = tokenizer.NewSimpleTokenizer()
	}
	return m
}

// Execute executes the rate limiting middleware.
//
// Summary: Enforces rate limits on tool execution.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *tool.ExecutionRequest. The request object.
//   - next: tool.ExecutionFunc. The next function in the chain.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if the rate limit is exceeded or execution fails.
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
			if toolErr != nil {
				return nil, fmt.Errorf("failed to get rate limiter for tool %s: %w", req.ToolName, toolErr)
			}
			if err := m.checkLimit(ctx, toolLimiter, toolConfig, req); err != nil {
				m.recordMetrics(serviceID, "tool", "blocked")
				return nil, fmt.Errorf("rate limit exceeded for tool %s: %w", req.ToolName, err)
			}
			m.recordMetrics(serviceID, "tool", "allowed")
			// If we checked tool limit, we return early (override logic).
			return next(ctx, req)
		}
	}

	// If no tool limit applied, check service limit
	if !appliedLimit && serviceRateLimitConfig != nil && serviceRateLimitConfig.GetIsEnabled() {
		serviceLimiter, serviceErr := m.getLimiter(ctx, serviceID, "service", serviceRateLimitConfig)
		if serviceErr != nil {
			return nil, fmt.Errorf("failed to get rate limiter for service %s: %w", serviceInfo.Name, serviceErr)
		}
		if err := m.checkLimit(ctx, serviceLimiter, serviceRateLimitConfig, req); err != nil {
			m.recordMetrics(serviceID, "service", "blocked")
			return nil, fmt.Errorf("rate limit exceeded for service %s: %w", serviceInfo.Name, err)
		}
		m.recordMetrics(serviceID, "service", "allowed")
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
		// Fail secure on error
		return fmt.Errorf("rate limit check failed: %w", err)
	}
	if !allowed {
		return fmt.Errorf("limit exceeded")
	}
	return nil
}

func (m *RateLimitMiddleware) recordMetrics(serviceID, limitType, status string) {
	metrics.IncrCounterWithLabels(metricRateLimitRequestsTotal, 1, []armonmetrics.Label{
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

	storageType := config.GetStorage()
	strategy, ok := m.strategies[storageType]
	if !ok {
		// Fallback to local if strategy not found or default to memory
		strategy = m.strategies[configv1.RateLimitConfig_STORAGE_MEMORY]
	}

	// Try to get from cache
	if val, found := m.limiters.Get(cacheKey); found {
		limiter := val.(Limiter)

		// We could try to verify if the limiter matches the strategy, but it's hard to do without type assertions on specific implementations.
		// For now, we assume that if the config changed significantly (storage type change), the strategy would be different.
		// But here we only have the generic Limiter interface.

		// Ideally the Strategy should handle validation/updating of existing limiter, but GetLimiter interface creates new one.
		// We can add "Validate" or "Update" to Strategy?
		// Or we just update the generic properties:
		limiter.Update(rps, burst)

		// However, for Redis, we need to check if connection changed.
		// This logic was previously inline.
		// To keep it clean, let's assume if we retrieved it, it's good, unless we want to do strict checking.
		// The previous logic checked `GetConfigHash()`.
		// Let's rely on the strategy to support checking?
		// Or simply: if it's RedisLimiter, we check.

		if rl, ok := limiter.(interface{ GetConfigHash() string }); ok && storageType == configv1.RateLimitConfig_STORAGE_REDIS && config.GetRedis() != nil {
			newConfigHash := config.GetRedis().GetAddress() + "|" + config.GetRedis().GetPassword() + "|" + strconv.Itoa(int(config.GetRedis().GetDb()))
			if rl.GetConfigHash() != newConfigHash {
				// Config changed, recreate
				goto Create
			}
		}

		return limiter, nil
	}

Create:
	// Create new limiter
	limiter, err := strategy.Create(ctx, serviceID, limitScopeKey, partitionKey, config)
	if err != nil {
		return nil, err
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
		if req, ok := ctx.Value(HTTPRequestContextKey).(*http.Request); ok {
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
