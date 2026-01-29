// Package service provides resilience patterns.

package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// UnaryClientInterceptor returns a new unary client interceptor that retries calls.
//
// retryConfig is the retryConfig.
//
// Returns the result.
func UnaryClientInterceptor(retryConfig *configv1.RetryConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		b := newBackoff(ctx, retryConfig)
		var err error
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			err = invoker(ctx, method, req, reply, cc, opts...)
			if !isRetryable(err) {
				return err
			}

			nextBackOff := b.NextBackOff()
			if nextBackOff == backoff.Stop {
				return err
			}

			slog.Warn("retrying call", "method", method, "error", err, "next_backoff", nextBackOff)

			timer := time.NewTimer(nextBackOff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}

func newBackoff(ctx context.Context, retryConfig *configv1.RetryConfig) backoff.BackOff {
	if retryConfig == nil {
		return &backoff.StopBackOff{}
	}

	b := &backoff.ExponentialBackOff{
		InitialInterval:     retryConfig.GetBaseBackoff().AsDuration(),
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         retryConfig.GetMaxBackoff().AsDuration(),
		MaxElapsedTime:      retryConfig.GetMaxElapsedTime().AsDuration(),
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	if b.InitialInterval == 0 {
		b.InitialInterval = 100 * time.Millisecond
	}
	b.Reset()
	retries := retryConfig.GetNumberOfRetries()
	if retries < 0 {
		retries = 0
	}
	return backoff.WithContext(backoff.WithMaxRetries(b, uint64(retries)), ctx) //nolint:gosec // safe cast
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch s.Code() {
	case codes.ResourceExhausted, codes.Unavailable, codes.Internal:
		return true
	default:
		return false
	}
}
