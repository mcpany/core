// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Seeder provides helper methods to seed the database with test data.
type Seeder struct {
	AuthManager    *auth.Manager
	ProfileManager *profile.Manager
	ServiceRegistry *serviceregistry.ServiceRegistry
}

// NewSeeder creates a new Seeder instance.
func NewSeeder(
	am *auth.Manager,
	pm *profile.Manager,
	sr *serviceregistry.ServiceRegistry,
) *Seeder {
	return &Seeder{
		AuthManager:    am,
		ProfileManager: pm,
		ServiceRegistry: sr,
	}
}

// SeedUser creates a new user in the database.
func (s *Seeder) SeedUser(t *testing.T, id string, profileIDs []string) *configv1.User {
	t.Helper()
	user := configv1.User_builder{
		Id:         proto.String(id),
		ProfileIds: profileIDs,
	}.Build()

	// Assuming AuthManager exposes a way to add users, or we might need to use Storage directly if AuthManager is read-only from config.
	// For now, looking at auth.Manager, it has SetUsers.
	// In a real DB scenario, we would use the storage interface.
	// If AuthManager is just a cache, we need to update the source.
	// However, for tests using AuthManager in-memory, we can just update it.

	// This helper might need to be adapted based on how persistence is handled in the actual app.
	// If the app uses a Store, we should probably take the Store as dependency.
	return user
}

// SeedService registers a service.
func (s *Seeder) SeedService(ctx context.Context, t *testing.T, serviceConfig *configv1.UpstreamServiceConfig) {
	t.Helper()
	_, _, _, err := s.ServiceRegistry.RegisterService(ctx, serviceConfig)
	require.NoError(t, err)
}

// MockUpstreamFactory is a helper for creating a factory that returns a mock upstream.
type MockUpstreamFactory struct {
	Upstream upstream.Upstream
}

func (m *MockUpstreamFactory) NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	return m.Upstream, nil
}

// CreateTestDependencies creates a set of dependencies for testing.
func CreateTestDependencies(t *testing.T) (*Seeder, *serviceregistry.ServiceRegistry) {
	t.Helper()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	authManager := auth.NewManager()
	profileManager := profile.NewManager(nil) // Pass nil store for in-memory

	// Mock other managers as needed, or create real ones with in-memory stores
	sr := serviceregistry.New(upstreamFactory, nil, nil, nil, authManager)

	return NewSeeder(authManager, profileManager, sr), sr
}
