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
		err := s.SaveGlobalSettings(ctx, globalSettings)
		assert.NoError(t, err)

		// Add a profile to test merging in Load
		profile := &configv1.ProfileDefinition{
			Name: proto.String("test-profile"),
		}
		err = s.SaveProfile(ctx, profile)
		assert.NoError(t, err)

		cfg, err := s.Load(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Len(t, cfg.UpstreamServices, 1)
		assert.Equal(t, "test-service", cfg.UpstreamServices[0].GetName())
		assert.NotNil(t, cfg.GlobalSettings)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, cfg.GlobalSettings.GetLogLevel())

		// Check if profiles are merged
		assert.Len(t, cfg.GlobalSettings.ProfileDefinitions, 1)
		assert.Equal(t, "test-profile", cfg.GlobalSettings.ProfileDefinitions[0].GetName())
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
		got, err := s2.GetGlobalSettings(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		// Default enum value is 0 (LOG_LEVEL_UNSPECIFIED)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_UNSPECIFIED, got.GetLogLevel())

		infoLevel := configv1.GlobalSettings_LOG_LEVEL_INFO
		settings := &configv1.GlobalSettings{
			LogLevel: &infoLevel,
		}
		err = s2.SaveGlobalSettings(ctx, settings)
		assert.NoError(t, err)

		got, err = s2.GetGlobalSettings(ctx)
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
		err := s3.SaveSecret(ctx, secret)
		assert.NoError(t, err)

		// Get
		got, err := s3.GetSecret(ctx, "sec-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "super-secret", got.GetValue())

		// Get Non-Existent
		got, err = s3.GetSecret(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s3.ListSecrets(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "sec-1", list[0].GetId())

		// Delete
		err = s3.DeleteSecret(ctx, "sec-1")
		assert.NoError(t, err)

		got, err = s3.GetSecret(ctx, "sec-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s3.ListSecrets(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Users", func(t *testing.T) {
		s4 := NewStore()
		user := &configv1.User{
			Id:   proto.String("user-1"),
			Roles: []string{"admin"},
		}

		// Create
		err := s4.CreateUser(ctx, user)
		assert.NoError(t, err)

		// Create Duplicate
		err = s4.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		// Create invalid
		err = s4.CreateUser(ctx, &configv1.User{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")

		// Get
		got, err := s4.GetUser(ctx, "user-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "user-1", got.GetId())
		assert.Contains(t, got.GetRoles(), "admin")

		// Get Non-Existent
		got, err = s4.GetUser(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s4.ListUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "user-1", list[0].GetId())

		// Update
		user.Roles = []string{"user"}
		err = s4.UpdateUser(ctx, user)
		assert.NoError(t, err)

		got, err = s4.GetUser(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "user", got.GetRoles()[0])

		// Update Non-Existent
		err = s4.UpdateUser(ctx, &configv1.User{Id: proto.String("non-existent")})
		assert.Error(t, err)

		// Delete
		err = s4.DeleteUser(ctx, "user-1")
		assert.NoError(t, err)

		got, err = s4.GetUser(ctx, "user-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Profiles", func(t *testing.T) {
		s5 := NewStore()
		profile := &configv1.ProfileDefinition{
			Name: proto.String("profile-1"),
		}

		// Save
		err := s5.SaveProfile(ctx, profile)
		assert.NoError(t, err)

		// Get
		got, err := s5.GetProfile(ctx, "profile-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "profile-1", got.GetName())

		// Get Non-Existent
		got, err = s5.GetProfile(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s5.ListProfiles(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)

		// Delete
		err = s5.DeleteProfile(ctx, "profile-1")
		assert.NoError(t, err)

		got, err = s5.GetProfile(ctx, "profile-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Service Collections", func(t *testing.T) {
		s6 := NewStore()
		collection := &configv1.UpstreamServiceCollectionShare{
			Name: proto.String("collection-1"),
		}

		// Save
		err := s6.SaveServiceCollection(ctx, collection)
		assert.NoError(t, err)

		// Get
		got, err := s6.GetServiceCollection(ctx, "collection-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "collection-1", got.GetName())

		// Get Non-Existent
		got, err = s6.GetServiceCollection(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s6.ListServiceCollections(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)

		// Delete
		err = s6.DeleteServiceCollection(ctx, "collection-1")
		assert.NoError(t, err)

		got, err = s6.GetServiceCollection(ctx, "collection-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Credentials", func(t *testing.T) {
		s7 := NewStore()
		cred := &configv1.Credential{
			Id: proto.String("cred-1"),
		}

		// Save
		err := s7.SaveCredential(ctx, cred)
		assert.NoError(t, err)

		// Get
		got, err := s7.GetCredential(ctx, "cred-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "cred-1", got.GetId())

		// Get Non-Existent
		got, err = s7.GetCredential(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s7.ListCredentials(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 1)

		// Delete
		err = s7.DeleteCredential(ctx, "cred-1")
		assert.NoError(t, err)

		got, err = s7.GetCredential(ctx, "cred-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Tokens", func(t *testing.T) {
		s8 := NewStore()
		token := &configv1.UserToken{
			UserId:    proto.String("user-1"),
			ServiceId: proto.String("service-1"),
			AccessToken:     proto.String("token-val"),
		}

		// Save
		err := s8.SaveToken(ctx, token)
		assert.NoError(t, err)

		// Get
		got, err := s8.GetToken(ctx, "user-1", "service-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "token-val", got.GetAccessToken())

		// Get Non-Existent
		got, err = s8.GetToken(ctx, "user-1", "service-2")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// Delete
		err = s8.DeleteToken(ctx, "user-1", "service-1")
		assert.NoError(t, err)

		got, err = s8.GetToken(ctx, "user-1", "service-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Close", func(t *testing.T) {
		err := s.Close()
		assert.NoError(t, err)
	})
}
