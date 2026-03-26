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

// MockStore ...
// Summary: MockStore
	Config *configv1.McpAnyServerConfig
	Err    error
}

// Load ...
// Summary: Load
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.Config, m.Err
}

// SaveService ...
// Summary: SaveService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.Err
}

// GetService ...
// Summary: GetService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, m.Err
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.Config != nil {
		return m.Config.GetUpstreamServices(), nil
	}
	return nil, m.Err
}

// DeleteService ...
// Summary: DeleteService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.Err
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// HasConfigSources ...
// Summary: HasConfigSources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// TestMultiStore ...
// Summary: TestMultiStore
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Run("MergeConfigs", func(t *testing.T) {
		s1 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					GlobalSettings: configv1.GlobalSettings_builder{
						ApiKey: proto.String("key1"),
					}.Build(),
				}.Build()
			}(),
		}
		s2 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					GlobalSettings: configv1.GlobalSettings_builder{
						McpListenAddress: proto.String(":8080"),
					}.Build(),
				}.Build()
			}(),
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key1", cfg.GetGlobalSettings().GetApiKey())
		assert.Equal(t, ":8080", cfg.GetGlobalSettings().GetMcpListenAddress())
	})

	t.Run("OverrideValues", func(t *testing.T) {
		s1 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					GlobalSettings: configv1.GlobalSettings_builder{
						ApiKey: proto.String("key1"),
					}.Build(),
				}.Build()
			}(),
		}
		s2 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					GlobalSettings: configv1.GlobalSettings_builder{
						ApiKey: proto.String("key2"),
					}.Build(),
				}.Build()
			}(),
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key2", cfg.GetGlobalSettings().GetApiKey())
	})

	t.Run("ErrorInStore", func(t *testing.T) {
		s1 := &MockStore{
			Config: configv1.McpAnyServerConfig_builder{}.Build(),
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
			Config: func() *configv1.McpAnyServerConfig {
				svc1 := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("svc1"),
				}.Build()
				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc1},
				}.Build()
			}(),
		}
		s2 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				svc2 := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("svc2"),
				}.Build()
				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc2},
				}.Build()
			}(),
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Len(t, cfg.GetUpstreamServices(), 2)
		assert.Equal(t, "svc1", cfg.GetUpstreamServices()[0].GetName())
		assert.Equal(t, "svc2", cfg.GetUpstreamServices()[1].GetName())
	})

	t.Run("NilConfigIgnored", func(t *testing.T) {
		s1 := &MockStore{
			Config: nil, // Should be ignored
		}
		s2 := &MockStore{
			Config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					GlobalSettings: configv1.GlobalSettings_builder{
						ApiKey: proto.String("key2"),
					}.Build(),
				}.Build()
			}(),
		}

		ms := NewMultiStore(s1, s2)
		cfg, err := ms.Load(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, "key2", cfg.GetGlobalSettings().GetApiKey())
	})
}
