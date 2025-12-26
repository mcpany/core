// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestProviderFactoryLogic tests the logic inside the default provider factory
// in NewCachingMiddleware.
func TestDefaultProviderFactory(t *testing.T) {
	mw := NewCachingMiddleware(&mockToolManagerForFactory{})
	factory := mw.providerFactory

	t.Run("OpenAI Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Openai{
				Openai: &configv1.OpenAIEmbeddingProviderConfig{
					Model: proto.String("gpt-4"),
				},
			},
		}

		provider, err := factory(conf, "test-key")
		require.NoError(t, err)
		assert.IsType(t, &OpenAIEmbeddingProvider{}, provider)
		p := provider.(*OpenAIEmbeddingProvider)
		assert.Equal(t, "test-key", p.apiKey)
		assert.Equal(t, "gpt-4", p.model)
	})

	t.Run("Ollama Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Ollama{
				Ollama: &configv1.OllamaEmbeddingProviderConfig{
					BaseUrl: proto.String("http://localhost:11434"),
					Model:   proto.String("llama2"),
				},
			},
		}

		provider, err := factory(conf, "")
		require.NoError(t, err)
		assert.IsType(t, &OllamaEmbeddingProvider{}, provider)
		p := provider.(*OllamaEmbeddingProvider)
		assert.Equal(t, "http://localhost:11434", p.baseURL)
		assert.Equal(t, "llama2", p.model)
	})

	t.Run("HTTP Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Http{
				Http: &configv1.HttpEmbeddingProviderConfig{
					Url:              proto.String("http://example.com/embed"),
					Headers:          map[string]string{"Authorization": "Bearer token"},
					BodyTemplate:     proto.String(`{"input": "{{.Input}}"}`),
					ResponseJsonPath: proto.String("data.embedding"),
				},
			},
		}

		provider, err := factory(conf, "")
		require.NoError(t, err)
		assert.IsType(t, &HttpEmbeddingProvider{}, provider)
		p := provider.(*HttpEmbeddingProvider)
		assert.Equal(t, "http://example.com/embed", p.url)
	})

	t.Run("Legacy OpenAI Config", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			Provider: proto.String("openai"),
			Model:    proto.String("gpt-3.5-turbo"),
		}

		provider, err := factory(conf, "legacy-key")
		require.NoError(t, err)
		assert.IsType(t, &OpenAIEmbeddingProvider{}, provider)
		p := provider.(*OpenAIEmbeddingProvider)
		assert.Equal(t, "legacy-key", p.apiKey)
		assert.Equal(t, "gpt-3.5-turbo", p.model)
	})

	t.Run("Unknown Provider", func(t *testing.T) {
		conf := &configv1.SemanticCacheConfig{
			Provider: proto.String("unknown"),
		}

		_, err := factory(conf, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider")
	})
}

// Minimal mock needed for NewCachingMiddleware
type mockToolManagerForFactory struct{}

func (m *mockToolManagerForFactory) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockToolManagerForFactory) AddTool(_ tool.Tool) error                { return nil }
func (m *mockToolManagerForFactory) GetTool(_ string) (tool.Tool, bool)       { return nil, false }
func (m *mockToolManagerForFactory) ListTools() []tool.Tool                   { return nil }
func (m *mockToolManagerForFactory) ListServices() []*tool.ServiceInfo        { return nil }
func (m *mockToolManagerForFactory) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *mockToolManagerForFactory) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (interface{}, error) {
	return nil, nil
}
func (m *mockToolManagerForFactory) SetMCPServer(_ tool.MCPServerProvider)        {}
func (m *mockToolManagerForFactory) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManagerForFactory) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *mockToolManagerForFactory) ClearToolsForService(_ string)                {}
