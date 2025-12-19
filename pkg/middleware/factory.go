package middleware

import (
	"fmt"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// CreateMiddleware creates a middleware instance from configuration.
func CreateMiddleware(serviceID string, config *configv1.Middleware) (Middleware, error) {
	switch c := config.MiddlewareConfig.(type) {
	case *configv1.Middleware_RateLimit:
		return NewRateLimitMiddleware(serviceID, c.RateLimit)
	case *configv1.Middleware_Cache:
		return NewCachingMiddleware(c.Cache), nil
	case *configv1.Middleware_CallPolicy:
		return NewCallPolicyMiddleware(serviceID, c.CallPolicy), nil
	case *configv1.Middleware_Audit:
		return NewAuditMiddleware(c.Audit)
	case *configv1.Middleware_Authentication:
		return NewAuthenticationMiddleware(c.Authentication), nil
	default:
		return nil, fmt.Errorf("unknown middleware type: %T", config.MiddlewareConfig)
	}
}
