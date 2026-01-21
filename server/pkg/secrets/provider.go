// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"
	"errors"
)

var (
	// ErrSecretNotFound is returned when a secret is not found.
	ErrSecretNotFound = errors.New("secret not found")
)

// Provider defines the interface for a secret storage provider.
type Provider interface {
	// GetSecret retrieves the secret value for the given ID.
	GetSecret(ctx context.Context, id string) (string, error)

	// RotateSecret rotates the secret for the given ID and returns the new version.
	// Some providers might not support manual rotation.
	RotateSecret(ctx context.Context, id string) (string, error)
}

// MockProvider is a mock implementation of Provider for testing.
type MockProvider struct {
	Secrets map[string]string
}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		Secrets: make(map[string]string),
	}
}

// GetSecret retrieves a secret from the mock store.
func (m *MockProvider) GetSecret(_ context.Context, id string) (string, error) {
	if val, ok := m.Secrets[id]; ok {
		return val, nil
	}
	return "", ErrSecretNotFound
}

// RotateSecret updates a secret in the mock store with a new value (simulated rotation).
func (m *MockProvider) RotateSecret(_ context.Context, id string) (string, error) {
	if _, ok := m.Secrets[id]; !ok {
		return "", ErrSecretNotFound
	}
	// Simulate rotation by appending "_rotated"
	newValue := m.Secrets[id] + "_rotated"
	m.Secrets[id] = newValue
	return newValue, nil
}
