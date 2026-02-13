// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SSRF_Protection(t *testing.T) {
	// Disable Dangerous Allow Local IPs if set, to ensure we test SSRF protection
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")

	// We must re-initialize the global httpClient used by config loader because it is initialized at package init time,
	// potentially before our t.Setenv took effect if run in the same process as other tests.
	// We save and restore the original client to avoid side effects.
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	client := util.NewSafeHTTPClient()
	client.Timeout = 5 * time.Second
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	httpClient = client

	// This test confirms that the SSRF protection blocks access to 127.0.0.1.
	// It relies on the default configuration of the secure http client.

	// Start a local HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte("global_settings:\n  log_level: INFO\n"))
	}))
	defer ts.Close()

	// Append a path with extension so NewEngine detects it as YAML
	configURL := ts.URL + "/config.yaml"

	fs := afero.NewMemMapFs()
	store := NewFileStore(fs, []string{configURL})

	// Attempt to load from 127.0.0.1
	_, err := store.Load(context.Background())

	// Assert that we get an error due to SSRF protection
	require.Error(t, err, "Should be blocked by SSRF protection")
	assert.Contains(t, err.Error(), "ssrf attempt blocked", "Error should indicate SSRF block")
}
