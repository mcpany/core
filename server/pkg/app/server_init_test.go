// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

type MockStore struct {
	SaveUserFunc           func(ctx context.Context, user *configv1.User) error
	ListUsersFunc          func(ctx context.Context) ([]*configv1.User, error)
	GetGlobalSettingsFunc  func(ctx context.Context) (*configv1.GlobalSettings, error)
	SaveGlobalSettingsFunc func(ctx context.Context, settings *configv1.GlobalSettings) error
	CreateUserFunc         func(ctx context.Context, user *configv1.User) error
}

func (m *MockStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return nil, nil
}

func (m *MockStore) HasConfigSources() bool {
	return false
}

func (m *MockStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return nil
}

func (m *MockStore) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}

func (m *MockStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}

func (m *MockStore) DeleteService(ctx context.Context, name string) error {
	return nil
}

func (m *MockStore) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	if m.GetGlobalSettingsFunc != nil {
		return m.GetGlobalSettingsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	if m.SaveGlobalSettingsFunc != nil {
		return m.SaveGlobalSettingsFunc(ctx, settings)
	}
	return nil
}

func (m *MockStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	return nil, nil
}

func (m *MockStore) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	return nil, nil
}

func (m *MockStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	return nil
}

func (m *MockStore) DeleteSecret(ctx context.Context, id string) error {
	return nil
}

func (m *MockStore) CreateUser(ctx context.Context, user *configv1.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockStore) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	return nil, nil
}

func (m *MockStore) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) UpdateUser(ctx context.Context, user *configv1.User) error {
	return nil
}

func (m *MockStore) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *MockStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	return nil, nil
}

func (m *MockStore) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	return nil, nil
}

func (m *MockStore) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	return nil
}

func (m *MockStore) DeleteProfile(ctx context.Context, name string) error {
	return nil
}

func (m *MockStore) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	return nil, nil
}

func (m *MockStore) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	return nil, nil
}

func (m *MockStore) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error {
	return nil
}

func (m *MockStore) DeleteServiceCollection(ctx context.Context, name string) error {
	return nil
}

func (m *MockStore) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	return nil
}

func (m *MockStore) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	return nil, nil
}

func (m *MockStore) DeleteToken(ctx context.Context, userID, serviceID string) error {
	return nil
}

func (m *MockStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	return nil, nil
}

func (m *MockStore) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	return nil, nil
}

func (m *MockStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	return nil
}

func (m *MockStore) DeleteCredential(ctx context.Context, id string) error {
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

func (m *MockStore) QueryLogs(ctx context.Context, filter logging.LogFilter) ([]logging.LogEntry, int, error) {
	return nil, 0, nil
}
