// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/stretchr/testify/require"
)

// SeedServices seeds the storage with the provided services.
func SeedServices(t *testing.T, ctx context.Context, store storage.Storage, services []*configv1.UpstreamServiceConfig) {
	for _, svc := range services {
		err := store.SaveService(ctx, svc)
		require.NoError(t, err, "failed to seed service: %s", svc.GetName())
	}
}

// SeedUsers seeds the storage with the provided users.
func SeedUsers(t *testing.T, ctx context.Context, store storage.Storage, users []*configv1.User) {
	for _, user := range users {
		err := store.CreateUser(ctx, user)
		require.NoError(t, err, "failed to seed user: %s", user.GetId())
	}
}
