// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type MockStore struct {
	Config *configv1.McpAnyServerConfig
	Err    error
}

func (m *MockStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return m.Config, m.Err
}

func (m *MockStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return m.Err
}

func (m *MockStore) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	return nil, m.Err
}

func (m *MockStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if m.Config != nil {
		return m.Config.UpstreamServices, nil
	}
	return nil, m.Err
}

func (m *MockStore) DeleteService(ctx context.Context, name string) error {
	return m.Err
}

func (m *MockStore) Close() error {
	return nil
}

func (m *MockStore) HasConfigSources() bool {
	return true
}

func TestMultiStore(t *testing.T) {
	t.Run("MergeConfigs", func(t *testing.T) {
		s1 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: proto.String("key1"),
				},
			},
		}
		s2 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					McpListenAddress: proto.String(":8080"),
				},
			},
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key1", cfg.GetGlobalSettings().GetApiKey())
		assert.Equal(t, ":8080", cfg.GetGlobalSettings().GetMcpListenAddress())
	})

	t.Run("OverrideValues", func(t *testing.T) {
		s1 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: proto.String("key1"),
				},
			},
		}
		s2 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: proto.String("key2"),
				},
			},
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key2", cfg.GetGlobalSettings().GetApiKey())
	})

	t.Run("ErrorInStore", func(t *testing.T) {
		s1 := &MockStore{
			Config: &configv1.McpAnyServerConfig{},
		}
		s2 := &MockStore{
			Err: errors.New("load error"),
		}

		ms := NewMultiStore(s1, s2)
		_, err := ms.Load(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "load error", err.Error())
	})

	t.Run("MergeLists", func(t *testing.T) {
		s1 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{Name: proto.String("svc1")},
				},
			},
		}
		s2 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{Name: proto.String("svc2")},
				},
			},
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Len(t, cfg.UpstreamServices, 2)
		assert.Equal(t, "svc1", cfg.UpstreamServices[0].GetName())
		assert.Equal(t, "svc2", cfg.UpstreamServices[1].GetName())
	})

	t.Run("NilConfigIgnored", func(t *testing.T) {
		s1 := &MockStore{
			Config: nil, // Should be ignored
		}
		s2 := &MockStore{
			Config: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: proto.String("key2"),
				},
			},
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key2", cfg.GetGlobalSettings().GetApiKey())
	})
}
