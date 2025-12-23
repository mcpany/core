// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type MockSampler struct {
	Calls  int
	Result *mcp.CreateMessageResult
	Error  error
}

func (m *MockSampler) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	m.Calls++
	return m.Result, m.Error
}

func TestCachingSampler(t *testing.T) {
	ctx := context.Background()
	params := &mcp.CreateMessageParams{
		Messages: []*mcp.SamplingMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Hello"},
			},
		},
	}
	result := &mcp.CreateMessageResult{
		Role:    "assistant",
		Content: &mcp.TextContent{Text: "Hi there"},
		Model:   "test-model",
	}

	t.Run("Disabled", func(t *testing.T) {
		mock := &MockSampler{Result: result}
		config := &configv1.CacheConfig{IsEnabled: proto.Bool(false)}
		sampler := NewCachingSampler(mock, config)

		// First call
		res1, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, result, res1)
		assert.Equal(t, 1, mock.Calls)

		// Second call
		res2, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, result, res2)
		assert.Equal(t, 2, mock.Calls)
	})

	t.Run("Enabled_Hit", func(t *testing.T) {
		mock := &MockSampler{Result: result}
		config := &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Minute),
		}
		sampler := NewCachingSampler(mock, config)

		// First call (Miss)
		res1, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, result, res1)
		assert.Equal(t, 1, mock.Calls)

		// Second call (Hit)
		res2, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, result, res2)
		assert.Equal(t, 1, mock.Calls)
	})

	t.Run("Enabled_DifferentParams", func(t *testing.T) {
		mock := &MockSampler{Result: result}
		config := &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Minute),
		}
		sampler := NewCachingSampler(mock, config)

		// First call
		_, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.Calls)

		// Second call (Different params)
		params2 := &mcp.CreateMessageParams{
			Messages: []*mcp.SamplingMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Different"},
				},
			},
		}
		_, err = sampler.CreateMessage(ctx, params2)
		assert.NoError(t, err)
		assert.Equal(t, 2, mock.Calls)
	})

	t.Run("Enabled_Expiration", func(t *testing.T) {
		mock := &MockSampler{Result: result}
		// Very short TTL
		config := &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Millisecond),
		}
		sampler := NewCachingSampler(mock, config)

		// First call
		_, err := sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.Calls)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Second call (Miss due to expiration)
		_, err = sampler.CreateMessage(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, 2, mock.Calls)
	})

	t.Run("Skip_IncludeContext", func(t *testing.T) {
		mock := &MockSampler{Result: result}
		config := &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Minute),
		}
		sampler := NewCachingSampler(mock, config)

		paramsWithContext := &mcp.CreateMessageParams{
			Messages: []*mcp.SamplingMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Context dependent"},
				},
			},
			IncludeContext: "allServers",
		}

		// First call
		_, err := sampler.CreateMessage(ctx, paramsWithContext)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.Calls)

		// Second call (Should SKIP cache, so call underlying again)
		_, err = sampler.CreateMessage(ctx, paramsWithContext)
		assert.NoError(t, err)
		assert.Equal(t, 2, mock.Calls)
	})
}
