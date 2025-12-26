// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCachingMiddleware_ProviderFactory(t *testing.T) {
	tm := new(MockToolManager)
	mw := NewCachingMiddleware(tm)

	t.Run("OpenAI Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Openai{
				Openai: &configv1.OpenAIEmbeddingProviderConfig{
					Model: proto.String("text-embedding-3-small"),
				},
			},
		}
		provider, err := mw.providerFactory(conf, "api-key")
		require.NoError(t, err)
		assert.IsType(t, &OpenAIEmbeddingProvider{}, provider)
		p := provider.(*OpenAIEmbeddingProvider)
		assert.Equal(t, "api-key", p.apiKey)
		assert.Equal(t, "text-embedding-3-small", p.model)
	})

	t.Run("Ollama Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Ollama{
				Ollama: &configv1.OllamaEmbeddingProviderConfig{
					BaseUrl: proto.String("http://localhost:11434"),
					Model:   proto.String("nomic-embed-text"),
				},
			},
		}
		provider, err := mw.providerFactory(conf, "")
		require.NoError(t, err)
		assert.IsType(t, &OllamaEmbeddingProvider{}, provider)
		p := provider.(*OllamaEmbeddingProvider)
		assert.Equal(t, "http://localhost:11434", p.baseURL)
		assert.Equal(t, "nomic-embed-text", p.model)
	})

	t.Run("HTTP Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Http{
				Http: &configv1.HttpEmbeddingProviderConfig{
					Url: proto.String("http://example.com"),
				},
			},
		}
		provider, err := mw.providerFactory(conf, "")
		require.NoError(t, err)
		assert.IsType(t, &HttpEmbeddingProvider{}, provider)
		p := provider.(*HttpEmbeddingProvider)
		assert.Equal(t, "http://example.com", p.url)
	})

	t.Run("Legacy OpenAI", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			Provider: proto.String("openai"),
			Model:    proto.String("text-embedding-ada-002"),
		}
		provider, err := mw.providerFactory(conf, "legacy-key")
		require.NoError(t, err)
		assert.IsType(t, &OpenAIEmbeddingProvider{}, provider)
		p := provider.(*OpenAIEmbeddingProvider)
		assert.Equal(t, "legacy-key", p.apiKey)
		assert.Equal(t, "text-embedding-ada-002", p.model)
	})

	t.Run("Unknown Provider", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			Provider: proto.String("unknown"),
		}
		provider, err := mw.providerFactory(conf, "")
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unknown provider")
	})
}

func TestCachingMiddleware_Execute_ToolNotFound(t *testing.T) {
	tm := new(MockToolManager)
	tm.On("GetTool", "unknown-tool").Return(nil, false)
	mw := NewCachingMiddleware(tm)

	req := &tool.ExecutionRequest{
		ToolName: "unknown-tool",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "ok", nil
	}

	res, err := mw.Execute(context.Background(), req, next)
	require.NoError(t, err)
	assert.Equal(t, "ok", res)
	assert.True(t, nextCalled)
	tm.AssertExpectations(t)
}

func TestCachingMiddleware_ExecuteSemantic_FactoryError(t *testing.T) {
	tm := new(MockToolManager)
	mw := NewCachingMiddleware(tm)

	// Force factory error
	mw.providerFactory = func(conf *configv1.SemanticCacheConfig, apiKey string) (EmbeddingProvider, error) {
		return nil, errors.New("factory error")
	}

	testTool := new(MockTool)
	testTool.On("Tool").Return(&v1.Tool{Name: proto.String("test-tool"), ServiceId: proto.String("test-service")})
	// We need Tool().GetName() if logging uses it via tool.Tool.GetName() which doesn't exist, it uses t.Tool().GetName()
	// The implementation calls t.Tool().GetName()

	ctx := tool.NewContextWithTool(context.Background(), testTool)

	cacheConfig := &configv1.CacheConfig{
		Strategy: proto.String("semantic"),
		SemanticConfig: &configv1.SemanticCacheConfig{
			Provider: proto.String("openai"),
		},
	}

	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "next", nil
	}

	res, err := mw.executeSemantic(ctx, req, next, testTool, cacheConfig, nil)
	require.NoError(t, err)
	assert.Equal(t, "next", res)
	assert.True(t, nextCalled)
}

func TestCachingMiddleware_ExecuteSemantic_VectorStore_SQLite(t *testing.T) {
	// Test the path where SQLite vector store is created
	tm := new(MockToolManager)
	mw := NewCachingMiddleware(tm)

	// Mock factory
	mw.providerFactory = func(conf *configv1.SemanticCacheConfig, apiKey string) (EmbeddingProvider, error) {
		return &MockEmbeddingProviderInternal{}, nil
	}

	testTool := new(MockTool)
	testTool.On("Tool").Return(&v1.Tool{Name: proto.String("test-tool"), ServiceId: proto.String("test-service")})

	ctx := tool.NewContextWithTool(context.Background(), testTool)

	// Create a temp file for sqlite
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cacheConfig := &configv1.CacheConfig{
		Strategy: proto.String("semantic"),
		SemanticConfig: &configv1.SemanticCacheConfig{
			Provider:        proto.String("openai"),
			PersistencePath: proto.String(dbPath),
		},
		Ttl: durationpb.New(time.Hour),
	}

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte("input"),
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "next", nil
	}

	// First call should initialize cache with SQLite store
	res, err := mw.executeSemantic(ctx, req, next, testTool, cacheConfig, nil)
	require.NoError(t, err)
	assert.Equal(t, "next", res)

	// Verify that cache entry exists and is using SQLite (implicit check)
	val, ok := mw.semanticCaches.Load("test-service")
	assert.True(t, ok)
	semCache := val.(*SemanticCache)
	assert.IsType(t, &SQLiteVectorStore{}, semCache.store)
}

type MockEmbeddingProviderInternal struct{}

func (m *MockEmbeddingProviderInternal) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}
