// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestHTTPCheckSSRF_Fixed verifies that the implementation blocks
// connections to loopback addresses.
func TestHTTPCheckSSRF_Fixed(t *testing.T) {
	// Store original env vars
	origLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	origDangerous := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	origPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")

	// Ensure loopback resources are NOT allowed by default for this test
	// This mitigates environment pollution where this might be set to true
	_ = os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	_ = os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	_ = os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")

	defer func() {
		// Restore env vars
		if origLoopback != "" {
			_ = os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", origLoopback)
		}
		if origDangerous != "" {
			_ = os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", origDangerous)
		}
		if origPrivate != "" {
			_ = os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", origPrivate)
		}
	}()

	// Start a local server on loopback
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure health check to point to this local server
	hc := &configv1.HttpHealthCheck{
		Url:          proto.String(server.URL),
		ExpectedCode: proto.Int32(200),
		Timeout:      durationpb.New(2 * time.Second),
	}

	// This should FAIL (return error) because we have fixed the SSRF.
	err := httpCheckFunc(context.Background(), "", hc)
	if err == nil {
		t.Logf("URL: %s", server.URL)
		t.Fatal("Expected error (SSRF blocked), got nil")
	}
	assert.Contains(t, err.Error(), "blocked", "Error should mention blocked")
}
