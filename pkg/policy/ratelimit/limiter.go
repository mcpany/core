package ratelimit

import (
	"golang.org/x/time/rate"
)

// Limiter defines the interface for a rate limiter.
type Limiter interface {
	// Allow reports whether an event may happen now.
	Allow() bool
}

// inMemoryLimiter is an in-memory implementation of the Limiter interface.
type inMemoryLimiter struct {
	limiter *rate.Limiter
}

// NewInMemoryLimiter creates a new in-memory rate limiter.
func NewInMemoryLimiter(requestsPerSecond float64, burst int) Limiter {
	return &inMemoryLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
	}
}

// Allow checks if a request is allowed based on the rate limit.
func (l *inMemoryLimiter) Allow() bool {
	return l.limiter.Allow()
}
