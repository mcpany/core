// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// MockStorage is a mock implementation of storage.Storage for testing purposes.
type MockStorage struct {
	GetCredentialFunc func(ctx context.Context, id string) (*configv1.Credential, error)
	GetServiceFunc    func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)
	SaveCredentialFunc func(ctx context.Context, cred *configv1.Credential) error
	SaveTokenFunc      func(ctx context.Context, token *configv1.UserToken) error
}

func (m *MockStorage) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return nil, nil
}

func (m *MockStorage) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return nil
}

func (m *MockStorage) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	if m.GetServiceFunc != nil {
		return m.GetServiceFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockStorage) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}

func (m *MockStorage) DeleteService(ctx context.Context, name string) error {
	return nil
}

func (m *MockStorage) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	return nil, nil
}

func (m *MockStorage) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	return nil
}

func (m *MockStorage) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	return nil, nil
}

func (m *MockStorage) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	return nil, nil
}

func (m *MockStorage) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	return nil
}

func (m *MockStorage) DeleteSecret(ctx context.Context, id string) error {
	return nil
}

func (m *MockStorage) CreateUser(ctx context.Context, user *configv1.User) error {
	return nil
}

func (m *MockStorage) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	return nil, nil
}

func (m *MockStorage) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	return nil, nil
}

func (m *MockStorage) UpdateUser(ctx context.Context, user *configv1.User) error {
	return nil
}

func (m *MockStorage) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *MockStorage) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	return nil, nil
}

func (m *MockStorage) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	return nil, nil
}

func (m *MockStorage) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	return nil
}

func (m *MockStorage) DeleteProfile(ctx context.Context, name string) error {
	return nil
}

func (m *MockStorage) ListServiceCollections(ctx context.Context) ([]*configv1.UpstreamServiceCollectionShare, error) {
	return nil, nil
}

func (m *MockStorage) GetServiceCollection(ctx context.Context, name string) (*configv1.UpstreamServiceCollectionShare, error) {
	return nil, nil
}

func (m *MockStorage) SaveServiceCollection(ctx context.Context, collection *configv1.UpstreamServiceCollectionShare) error {
	return nil
}

func (m *MockStorage) DeleteServiceCollection(ctx context.Context, name string) error {
	return nil
}

func (m *MockStorage) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	if m.SaveTokenFunc != nil {
		return m.SaveTokenFunc(ctx, token)
	}
	return nil
}

func (m *MockStorage) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	return nil, nil
}

func (m *MockStorage) DeleteToken(ctx context.Context, userID, serviceID string) error {
	return nil
}

func (m *MockStorage) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	return nil, nil
}

func (m *MockStorage) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	if m.GetCredentialFunc != nil {
		return m.GetCredentialFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStorage) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	if m.SaveCredentialFunc != nil {
		return m.SaveCredentialFunc(ctx, cred)
	}
	return nil
}

func (m *MockStorage) DeleteCredential(ctx context.Context, id string) error {
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}
