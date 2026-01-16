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
		err := s.SaveGlobalSettings(context.Background(), globalSettings)
		assert.NoError(t, err)

		// Set a profile as well to test integration in Load
		profile := &configv1.ProfileDefinition{
			Name: proto.String("test-profile"),
		}
		err = s.SaveProfile(context.Background(), profile)
		assert.NoError(t, err)

		cfg, err := s.Load(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Len(t, cfg.UpstreamServices, 1)
		assert.Equal(t, "test-service", cfg.UpstreamServices[0].GetName())
		assert.NotNil(t, cfg.GlobalSettings)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, cfg.GlobalSettings.GetLogLevel())
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
		got, err := s2.GetGlobalSettings(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, got)
		// Default enum value is 0 (LOG_LEVEL_UNSPECIFIED)
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_UNSPECIFIED, got.GetLogLevel())

		infoLevel := configv1.GlobalSettings_LOG_LEVEL_INFO
		settings := &configv1.GlobalSettings{
			LogLevel: &infoLevel,
		}
		err = s2.SaveGlobalSettings(context.Background(), settings)
		assert.NoError(t, err)

		got, err = s2.GetGlobalSettings(context.Background())
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
		err := s3.SaveSecret(context.Background(), secret)
		assert.NoError(t, err)

		// Get
		got, err := s3.GetSecret(context.Background(), "sec-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "super-secret", got.GetValue())

		// Get Non-Existent
		got, err = s3.GetSecret(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s3.ListSecrets(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "sec-1", list[0].GetId())

		// Delete
		err = s3.DeleteSecret(context.Background(), "sec-1")
		assert.NoError(t, err)

		got, err = s3.GetSecret(context.Background(), "sec-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s3.ListSecrets(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Users", func(t *testing.T) {
		s4 := NewStore()
		user := &configv1.User{
			Id:   proto.String("user-1"),
		}

		// Create
		err := s4.CreateUser(context.Background(), user)
		assert.NoError(t, err)

		// Create Duplicate
		err = s4.CreateUser(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")

		// Create Missing ID
		err = s4.CreateUser(context.Background(), &configv1.User{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")

		// Get
		got, err := s4.GetUser(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "user-1", got.GetId())

		// Get Non-Existent
		got, err = s4.GetUser(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s4.ListUsers(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "user-1", list[0].GetId())

		// Update
		// User only has Id in memory implementation usage effectively for now in test
		// but let's just update same user
		err = s4.UpdateUser(context.Background(), user)
		assert.NoError(t, err)

		got, err = s4.GetUser(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "user-1", got.GetId())

		// Update Non-Existent
		err = s4.UpdateUser(context.Background(), &configv1.User{Id: proto.String("non-existent")})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		// Delete
		err = s4.DeleteUser(context.Background(), "user-1")
		assert.NoError(t, err)

		got, err = s4.GetUser(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s4.ListUsers(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Profiles", func(t *testing.T) {
		s5 := NewStore()
		profile := &configv1.ProfileDefinition{
			Name: proto.String("profile-1"),
		}

		// Save
		err := s5.SaveProfile(context.Background(), profile)
		assert.NoError(t, err)

		// Get
		got, err := s5.GetProfile(context.Background(), "profile-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "profile-1", got.GetName())

		// Get Non-Existent
		got, err = s5.GetProfile(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s5.ListProfiles(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "profile-1", list[0].GetName())

		// Delete
		err = s5.DeleteProfile(context.Background(), "profile-1")
		assert.NoError(t, err)

		got, err = s5.GetProfile(context.Background(), "profile-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s5.ListProfiles(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Service Collections", func(t *testing.T) {
		s6 := NewStore()
		collection := &configv1.Collection{
			Name: proto.String("collection-1"),
			Description: proto.String("Test Collection"),
		}

		// Save
		err := s6.SaveServiceCollection(context.Background(), collection)
		assert.NoError(t, err)

		// Get
		got, err := s6.GetServiceCollection(context.Background(), "collection-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "Test Collection", got.GetDescription())

		// Get Non-Existent
		got, err = s6.GetServiceCollection(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s6.ListServiceCollections(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "collection-1", list[0].GetName())

		// Delete
		err = s6.DeleteServiceCollection(context.Background(), "collection-1")
		assert.NoError(t, err)

		got, err = s6.GetServiceCollection(context.Background(), "collection-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s6.ListServiceCollections(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Credentials", func(t *testing.T) {
		s7 := NewStore()
		cred := &configv1.Credential{
			Id: proto.String("cred-1"),
			Name: proto.String("Test Credential"),
		}

		// Save
		err := s7.SaveCredential(context.Background(), cred)
		assert.NoError(t, err)

		// Get
		got, err := s7.GetCredential(context.Background(), "cred-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "Test Credential", got.GetName())

		// Get Non-Existent
		got, err = s7.GetCredential(context.Background(), "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// List
		list, err := s7.ListCredentials(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "cred-1", list[0].GetId())

		// Delete
		err = s7.DeleteCredential(context.Background(), "cred-1")
		assert.NoError(t, err)

		got, err = s7.GetCredential(context.Background(), "cred-1")
		assert.NoError(t, err)
		assert.Nil(t, got)

		list, err = s7.ListCredentials(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Tokens", func(t *testing.T) {
		s8 := NewStore()
		token := &configv1.UserToken{
			UserId:    proto.String("user-1"),
			ServiceId: proto.String("service-1"),
			AccessToken: proto.String("abc-123"),
		}

		// Save
		err := s8.SaveToken(context.Background(), token)
		assert.NoError(t, err)

		// Get
		got, err := s8.GetToken(context.Background(), "user-1", "service-1")
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "abc-123", got.GetAccessToken())

		// Get Non-Existent
		got, err = s8.GetToken(context.Background(), "user-1", "service-2")
		assert.NoError(t, err)
		assert.Nil(t, got)

		// Delete
		err = s8.DeleteToken(context.Background(), "user-1", "service-1")
		assert.NoError(t, err)

		got, err = s8.GetToken(context.Background(), "user-1", "service-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Close", func(t *testing.T) {
		err := s.Close()
		assert.NoError(t, err)
	})
}
