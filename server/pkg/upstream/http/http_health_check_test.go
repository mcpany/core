// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
)

// TestHTTPUpstream_CheckHealth_BeforeRegister ...
// Summary: TestHTTPUpstream_CheckHealth_BeforeRegister
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	pm := pool.NewManager()
	upstream := NewUpstream(pm)

	type HealthChecker interface {
		CheckHealth(ctx context.Context) error
	}

	hc, ok := upstream.(HealthChecker)
	assert.True(t, ok)

	err := hc.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no address configured")
}
