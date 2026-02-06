// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_ListCommunityServers(t *testing.T) {
	m := NewManager()
	servers, err := m.ListCommunityServers(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, servers)

	// Verify we have Postgres
	found := false
	for _, s := range servers {
		if s.Name == "PostgreSQL" {
			found = true
			assert.NotEmpty(t, s.ConfigurationSchema)
			assert.NotEmpty(t, s.Command)
			break
		}
	}
	assert.True(t, found, "PostgreSQL should be in the registry")
}

func TestManager_FindServerByName(t *testing.T) {
	m := NewManager()

	s, found := m.FindServerByName("postgresql")
	assert.True(t, found)
	assert.Equal(t, "PostgreSQL", s.Name)

	s, found = m.FindServerByName("github")
	assert.True(t, found)
	assert.Equal(t, "GitHub", s.Name)

	_, found = m.FindServerByName("UnknownServer")
	assert.False(t, found)
}
