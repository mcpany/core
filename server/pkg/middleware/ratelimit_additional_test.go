// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// Export countChars for testing purposes since it is private
// In Go, tests in the same package (middleware) can access private identifiers.
// Since this file is in package middleware, we can access countChars directly.

func TestCountChars(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name:     "string",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "slice of strings",
			input:    []interface{}{"hello", "world"},
			expected: 10,
		},
		{
			name:     "slice mixed",
			input:    []interface{}{"hello", 123},
			expected: 5 + 3, // "hello" + "123"
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key1": "value1", // 6 chars
				"key2": "val2",   // 4 chars
			},
			expected: 10,
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"key1": map[string]interface{}{
					"inner": "val", // 3 chars
				},
			},
			expected: 3,
		},
		{
			name:     "integer",
			input:    12345,
			expected: 5, // "12345"
		},
		{
			name:     "float",
			input:    12.34,
			expected: 5, // "12.34"
		},
		{
			name:     "nil",
			input:    nil,
			expected: 5, // "<nil>"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, countChars(tt.input))
		})
	}
}

func TestEstimateTokenCost(t *testing.T) {
	m := &RateLimitMiddleware{}

	tests := []struct {
		name     string
		req      *tool.ExecutionRequest
		expected int
	}{
		{
			name: "arguments present",
			req: &tool.ExecutionRequest{
				Arguments: map[string]interface{}{
					"arg1": "1234",      // 4 chars
					"arg2": "5678",      // 4 chars
					"arg3": "90123456", // 8 chars
				}, // Total 16 chars -> 4 tokens
			},
			expected: 4,
		},
		{
			name: "tool inputs json",
			req: &tool.ExecutionRequest{
				ToolInputs: json.RawMessage(`{"arg1": "1234", "arg2": "5678"}`), // 8 chars -> 2 tokens
			},
			expected: 2,
		},
		{
			name: "tool inputs invalid json",
			req: &tool.ExecutionRequest{
				ToolInputs: json.RawMessage(`{invalid`),
			},
			expected: 1, // Default
		},
		{
			name:     "empty",
			req:      &tool.ExecutionRequest{},
			expected: 1, // Minimum
		},
		{
			name: "arguments small",
			req: &tool.ExecutionRequest{
				Arguments: map[string]interface{}{
					"arg1": "1",
				},
			},
			expected: 1, // Minimum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, m.estimateTokenCost(tt.req))
		})
	}
}

func TestGetPartitionKey(t *testing.T) {
	m := &RateLimitMiddleware{}

	ctx := context.Background()

	t.Run("IP", func(t *testing.T) {
		ctxWithIP := util.ContextWithRemoteIP(ctx, "1.2.3.4")
		key := m.getPartitionKey(ctxWithIP, configv1.RateLimitConfig_KEY_BY_IP)
		assert.Equal(t, "ip:1.2.3.4", key)

		keyUnknown := m.getPartitionKey(ctx, configv1.RateLimitConfig_KEY_BY_IP)
		assert.Equal(t, "ip:unknown", keyUnknown)
	})

	t.Run("User ID", func(t *testing.T) {
		ctxWithUser := auth.ContextWithUser(ctx, "user123")
		key := m.getPartitionKey(ctxWithUser, configv1.RateLimitConfig_KEY_BY_USER_ID)
		assert.Equal(t, "user:user123", key)

		keyUnknown := m.getPartitionKey(ctx, configv1.RateLimitConfig_KEY_BY_USER_ID)
		assert.Equal(t, "user:anonymous", keyUnknown)
	})

	t.Run("API Key", func(t *testing.T) {
		ctxWithAPIKey := auth.ContextWithAPIKey(ctx, "apikey123")
		key := m.getPartitionKey(ctxWithAPIKey, configv1.RateLimitConfig_KEY_BY_API_KEY)

		assert.NotEmpty(t, key)
		assert.NotEqual(t, "apikey:apikey123", key)

		// Fallback not easily testable without constructing full HTTP request in context which uses private keys in some frameworks,
		// or if we can use http.Request in context.
		// The code checks: if req, ok := ctx.Value("http.request").(*http.Request); ok
		// "http.request" is a string key.
	})

	t.Run("Default", func(t *testing.T) {
		key := m.getPartitionKey(ctx, configv1.RateLimitConfig_KeyBy(999))
		assert.Equal(t, "", key)
	})
}
