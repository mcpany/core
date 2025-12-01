// Copyright (C) 2024 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	configv1 "github.com/mcpany/core/proto/config/v1"
)
func TestIsRetryable(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "NilError",
			err:      nil,
			expected: false,
		},
		{
			name:     "NonGRPCError",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "RetryableCodeResourceExhausted",
			err:      status.Error(codes.ResourceExhausted, "resource exhausted"),
			expected: true,
		},
		{
			name:     "RetryableCodeUnavailable",
			err:      status.Error(codes.Unavailable, "unavailable"),
			expected: true,
		},
		{
			name:     "RetryableCodeInternal",
			err:      status.Error(codes.Internal, "internal error"),
			expected: true,
		},
		{
			name:     "NonRetryableCode",
			err:      status.Error(codes.InvalidArgument, "invalid argument"),
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isRetryable(tc.err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestUnaryClientInterceptor_MaxElapsedTime(t *testing.T) {
	retries := int32(20)
	retryConfig := configv1.RetryConfig_builder{
		NumberOfRetries: &retries,
		BaseBackoff:     durationpb.New(10 * time.Millisecond),
		MaxBackoff:      durationpb.New(100 * time.Millisecond),
		MaxElapsedTime:  durationpb.New(100 * time.Millisecond),
	}.Build()

	interceptor := UnaryClientInterceptor(retryConfig)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Unavailable, "unavailable")
	}

	start := time.Now()
	err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
	elapsed := time.Since(start)

	t.Logf("Elapsed time: %s", elapsed)

	assert.Error(t, err)
	assert.InDelta(t, float64(100*time.Millisecond), float64(elapsed), float64(60*time.Millisecond))
}
