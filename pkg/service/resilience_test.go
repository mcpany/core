// Copyright (C) 2024 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
