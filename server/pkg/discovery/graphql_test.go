// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLProvider_Name(t *testing.T) {
	provider := &GraphQLProvider{}
	assert.Equal(t, "graphql", provider.Name())
}

func TestGraphQLProvider_Discover(t *testing.T) {
	provider := &GraphQLProvider{Endpoint: "http://localhost:8080/graphql"}

	svcs, err := provider.Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, svcs, 1)

	svc := svcs[0]
	assert.Equal(t, "Auto-discovered GraphQL", svc.GetName())
	assert.Equal(t, "http://localhost:8080/graphql", svc.GetGraphqlService().GetAddress())
	assert.Contains(t, svc.GetTags(), "graphql")
	assert.Contains(t, svc.GetTags(), "auto-discovered")
}
