// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIProvider_Name ...
// Summary: TestOpenAPIProvider_Name
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	provider := &OpenAPIProvider{}
	assert.Equal(t, "openapi", provider.Name())
}

// TestOpenAPIProvider_Discover ...
// Summary: TestOpenAPIProvider_Discover
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	provider := &OpenAPIProvider{Endpoint: "http://localhost:8080/openapi.json"}

	svcs, err := provider.Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, svcs, 1)

	svc := svcs[0]
	assert.Equal(t, "Auto-discovered OpenAPI", svc.GetName())
	assert.Equal(t, "http://localhost:8080/openapi.json", svc.GetOpenapiService().GetAddress())
	assert.Contains(t, svc.GetTags(), "openapi")
	assert.Contains(t, svc.GetTags(), "auto-discovered")
}
