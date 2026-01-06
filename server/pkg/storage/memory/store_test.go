// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMemoryStore(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	t.Run("Save and Get Service", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			Id:   proto.String("id-1"),
		}
		err := s.SaveService(ctx, svc)
		assert.NoError(t, err)

		got, err := s.GetService(ctx, "test-service")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "test-service", got.GetName())
		assert.Equal(t, "id-1", got.GetId())
	})

	t.Run("Get Non-Existent Service", func(t *testing.T) {
		got, err := s.GetService(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("List Services", func(t *testing.T) {
		list, err := s.ListServices(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "test-service", list[0].GetName())
	})

	t.Run("Load Config", func(t *testing.T) {
		// Set some global settings first to test loading them too
		debugLevel := configv1.GlobalSettings_LOG_LEVEL_DEBUG
		globalSettings := &configv1.GlobalSettings{
			LogLevel: &debugLevel,
		}
		err := s.SaveGlobalSettings(globalSettings)
		assert.NoError(t, err)

		cfg, err := s.Load(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Len(t, cfg.UpstreamServices, 1)
		assert.Equal(t, "test-service", cfg.UpstreamServices[0].GetName())
		assert.NotNil(t, cfg.GlobalSettings)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, cfg.GlobalSettings.GetLogLevel())
	})

	t.Run("Delete Service", func(t *testing.T) {
		err := s.DeleteService(ctx, "test-service")
		assert.NoError(t, err)

		got, err := s.GetService(ctx, "test-service")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err := s.ListServices(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Global Settings", func(t *testing.T) {
		// Initial state (should be empty but not nil if we follow implementation,
		// actually implementation returns empty struct if nil)
		s2 := NewStore()
		got, err := s2.GetGlobalSettings()
		assert.NoError(t, err)
		assert.NotNil(t, got)
		// Default enum value is 0 (LOG_LEVEL_UNSPECIFIED)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_UNSPECIFIED, got.GetLogLevel())

		infoLevel := configv1.GlobalSettings_LOG_LEVEL_INFO
		settings := &configv1.GlobalSettings{
			LogLevel: &infoLevel,
		}
		err = s2.SaveGlobalSettings(settings)
		assert.NoError(t, err)

		got, err = s2.GetGlobalSettings()
		assert.NoError(t, err)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, got.GetLogLevel())
	})

	t.Run("Secrets", func(t *testing.T) {
		s3 := NewStore()
		secret := &configv1.Secret{
			Id:    proto.String("sec-1"),
			Value: proto.String("super-secret"),
		}

		// Save
		err := s3.SaveSecret(secret)
		assert.NoError(t, err)

		// Get
		got, err := s3.GetSecret("sec-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "super-secret", got.GetValue())

		// Get Non-Existent
		got, err = s3.GetSecret("non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s3.ListSecrets()
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "sec-1", list[0].GetId())

		// Delete
		err = s3.DeleteSecret("sec-1")
		assert.NoError(t, err)

		got, err = s3.GetSecret("sec-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s3.ListSecrets()
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Close", func(t *testing.T) {
		err := s.Close()
		assert.NoError(t, err)
	})
}
