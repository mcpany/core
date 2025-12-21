// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"net/http"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPool_SSRFPrevention(t *testing.T) {
	// Initialize a pool with default config
	p, err := NewHTTPPool(1, 1, 10, &configv1.UpstreamServiceConfig{})
	require.NoError(t, err)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	clientWrapper, err := p.Get(context.Background())
	require.NoError(t, err)

	client := clientWrapper.Client

	// Try to connect to a link-local address (AWS metadata)
	// We use a small timeout because it might hang if it tries to connect
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 169.254.169.254 is a well-known link-local address used for cloud metadata services
	req, err := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/latest/meta-data/", nil)
	require.NoError(t, err)

	_, err = client.Do(req)

	// With the fix, this should fail with a specific error message "ssrf attempt blocked"
	if err == nil {
		t.Fatal("Expected SSRF attempt to be blocked, but request succeeded")
	}

	// Check if the error message indicates it was blocked by our SafeDialer
	// If the fix is NOT present, this assertion will likely fail (because it will be a timeout or network error)
	assert.Contains(t, err.Error(), "ssrf attempt blocked")
}
