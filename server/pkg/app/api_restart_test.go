// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/stretchr/testify/assert"
)

type MockStoreWithGet struct {
	GetFunc func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)
	DeleteFunc func(ctx context.Context, name string) error
	GetServiceCollectionFunc func(ctx context.Context, name string) (*configv1.Collection, error)
	DeleteServiceCollectionFunc func(ctx context.Context, name string) error
	ListServicesFunc func(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)
}

func (m *MockStoreWithGet) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockStoreWithGet) DeleteService(ctx context.Context, name string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, name)
	}
	return nil
}

func (m *MockStoreWithGet) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	if m.GetServiceCollectionFunc != nil {
		return m.GetServiceCollectionFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockStoreWithGet) DeleteServiceCollection(ctx context.Context, name string) error {
	if m.DeleteServiceCollectionFunc != nil {
		return m.DeleteServiceCollectionFunc(ctx, name)
	}
	return nil
}

func (m *MockStoreWithGet) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if m.ListServicesFunc != nil {
		return m.ListServicesFunc(ctx)
	}
	return nil, nil
}

// Implement other interface methods as no-ops
func (m *MockStoreWithGet) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) { return nil, nil }
func (m *MockStoreWithGet) HasConfigSources() bool { return false }
func (m *MockStoreWithGet) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error { return nil }
func (m *MockStoreWithGet) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) { return nil, nil }
func (m *MockStoreWithGet) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error { return nil }
func (m *MockStoreWithGet) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) { return nil, nil }
func (m *MockStoreWithGet) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) { return nil, nil }
func (m *MockStoreWithGet) SaveSecret(ctx context.Context, secret *configv1.Secret) error { return nil }
func (s *MockStoreWithGet) DeleteSecret(ctx context.Context, id string) error { return nil }
func (s *MockStoreWithGet) CreateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockStoreWithGet) GetUser(ctx context.Context, id string) (*configv1.User, error) { return nil, nil }
func (s *MockStoreWithGet) ListUsers(ctx context.Context) ([]*configv1.User, error) { return nil, nil }
func (s *MockStoreWithGet) UpdateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockStoreWithGet) DeleteUser(ctx context.Context, id string) error { return nil }
func (s *MockStoreWithGet) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) { return nil, nil }
func (s *MockStoreWithGet) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) { return nil, nil }
func (s *MockStoreWithGet) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error { return nil }
func (s *MockStoreWithGet) DeleteProfile(ctx context.Context, name string) error { return nil }
func (s *MockStoreWithGet) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) { return nil, nil }
func (s *MockStoreWithGet) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error { return nil }
func (s *MockStoreWithGet) SaveToken(ctx context.Context, token *configv1.UserToken) error { return nil }
func (s *MockStoreWithGet) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) { return nil, nil }
func (s *MockStoreWithGet) DeleteToken(ctx context.Context, userID, serviceID string) error { return nil }
func (s *MockStoreWithGet) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) { return nil, nil }
func (s *MockStoreWithGet) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) { return nil, nil }
func (s *MockStoreWithGet) SaveCredential(ctx context.Context, cred *configv1.Credential) error { return nil }
func (s *MockStoreWithGet) DeleteCredential(ctx context.Context, id string) error { return nil }
func (s *MockStoreWithGet) Close() error { return nil }
func (s *MockStoreWithGet) QueryLogs(ctx context.Context, filter logging.LogFilter) ([]logging.LogEntry, int, error) { return nil, 0, nil }

func TestHandleServiceRestart(t *testing.T) {
	app := &Application{}
	store := &MockStoreWithGet{}

	// Just verify it compiles and we can pass store
	_ = app.handleServiceRestart(nil, nil, "test", store)
}
