// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPool_IdleConnTimeout_Bug(t *testing.T) {
	// Reproduce bug where IdleConnTimeout is hardcoded to 90s instead of using the passed idleTimeout parameter.

	expectedTimeout := 5 * time.Second

	// Create a pool with 5s idle timeout
	p, err := NewHTTPPool(1, 1, expectedTimeout, configv1.UpstreamServiceConfig_builder{}.Build())
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// Type assert to the concrete type *httpPool to access private fields
	hp, ok := p.(*httpPool)
	require.True(t, ok, "Returned pool should be *httpPool")

	// Check the transport's IdleConnTimeout
	// Before fix: it will be 90s (hardcoded)
	// After fix: it should be 5s (expectedTimeout)
	assert.Equal(t, expectedTimeout, hp.transport.IdleConnTimeout, "IdleConnTimeout should match the configured idleTimeout")
}
