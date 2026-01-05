// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStore_Services(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	// Test SaveService
	name := "service1"
	svc := &configv1.UpstreamServiceConfig{
		Name: &name,
	}
	err := s.SaveService(ctx, svc)
	assert.NoError(t, err)

	// Test GetService
	got, err := s.GetService(ctx, name)
	assert.NoError(t, err)
	assert.Equal(t, svc.GetName(), got.GetName())
	assert.NotSame(t, svc, got) // Should be a clone

	// Test ListServices
	list, err := s.ListServices(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, svc.GetName(), list[0].GetName())

	// Test Load
	cfg, err := s.Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, cfg.UpstreamServices, 1)
	assert.Equal(t, svc.GetName(), cfg.UpstreamServices[0].GetName())

	// Test DeleteService
	err = s.DeleteService(ctx, name)
	assert.NoError(t, err)
	got, err = s.GetService(ctx, name)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestStore_GlobalSettings(t *testing.T) {
	s := NewStore()

	// Default empty settings
	got, err := s.GetGlobalSettings()
	assert.NoError(t, err)
	assert.NotNil(t, got)

	// Save settings
	settings := &configv1.GlobalSettings{
		McpListenAddress: proto.String("localhost:8080"),
	}
	err = s.SaveGlobalSettings(settings)
	assert.NoError(t, err)

	// Get settings
	got, err = s.GetGlobalSettings()
	assert.NoError(t, err)
	assert.Equal(t, settings.McpListenAddress, got.McpListenAddress)
	assert.NotSame(t, settings, got)

	// Verify Load includes global settings
	cfg, err := s.Load(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, settings.McpListenAddress, cfg.GlobalSettings.McpListenAddress)
}

func TestStore_Secrets(t *testing.T) {
	s := NewStore()

	id := "secret1"
	secret := &configv1.Secret{
		Id:    proto.String(id),
		Value: proto.String("value1"),
	}

	// SaveSecret
	err := s.SaveSecret(secret)
	assert.NoError(t, err)

	// GetSecret
	got, err := s.GetSecret(id)
	assert.NoError(t, err)
	assert.Equal(t, secret.Id, got.Id)
	assert.Equal(t, secret.Value, got.Value)

	// ListSecrets
	list, err := s.ListSecrets()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, secret.Id, list[0].Id)

	// DeleteSecret
	err = s.DeleteSecret(id)
	assert.NoError(t, err)
	got, err = s.GetSecret(id)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestStore_Close(t *testing.T) {
	s := NewStore()
	assert.NoError(t, s.Close())
}

func TestStore_ProtoClone(t *testing.T) {
	// Ensure that modifications to returned objects don't affect stored state
	s := NewStore()
	name := "svc"
	svc := &configv1.UpstreamServiceConfig{Name: &name}
	_ = s.SaveService(context.Background(), svc)

	got, _ := s.GetService(context.Background(), name)
	newName := "modified"
	got.Name = &newName

	got2, _ := s.GetService(context.Background(), name)
	assert.Equal(t, "svc", *got2.Name)
}

func TestStore_GlobalSettingsClone(t *testing.T) {
	s := NewStore()
	settings := &configv1.GlobalSettings{ApiKey: proto.String("initial")}
	_ = s.SaveGlobalSettings(settings)

	got, _ := s.GetGlobalSettings()
	got.ApiKey = proto.String("modified")

	got2, _ := s.GetGlobalSettings()
	assert.Equal(t, "initial", *got2.ApiKey)
}

func TestStore_SecretClone(t *testing.T) {
	s := NewStore()
	secret := &configv1.Secret{Id: proto.String("1"), Value: proto.String("initial")}
	_ = s.SaveSecret(secret)

	got, _ := s.GetSecret("1")
	got.Value = proto.String("modified")

	got2, _ := s.GetSecret("1")
	assert.Equal(t, "initial", *got2.Value)
}

func TestStore_LoadClone(t *testing.T) {
	s := NewStore()
	name := "svc"
	_ = s.SaveService(context.Background(), &configv1.UpstreamServiceConfig{Name: &name})
	_ = s.SaveGlobalSettings(&configv1.GlobalSettings{ApiKey: proto.String("key")})

	cfg, _ := s.Load(context.Background())
	cfg.GlobalSettings.ApiKey = proto.String("modified")
	newName := "modified"
	cfg.UpstreamServices[0].Name = &newName

	gotSvc, _ := s.GetService(context.Background(), "svc")
	assert.Equal(t, "svc", *gotSvc.Name)

	gotSettings, _ := s.GetGlobalSettings()
	assert.Equal(t, "key", *gotSettings.ApiKey)
}

// proto.Clone wrapper testing implicitly done above but ensuring interface works
func TestStore_Concurrency(t *testing.T) {
	// Simple race check
	s := NewStore()
	var wg sync.WaitGroup
	ctx := context.Background()

	name := "svc"
	s.SaveService(ctx, &configv1.UpstreamServiceConfig{Name: &name})

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.GetService(ctx, name)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			newName := "svc"
			s.SaveService(ctx, &configv1.UpstreamServiceConfig{Name: &newName})
		}
	}()
	wg.Wait()
}
