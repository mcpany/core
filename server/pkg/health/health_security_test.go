// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPCheckFunc_SSRF(t *testing.T) {
	// Ensure env vars are unset for this test to verify blocking behavior
	// Using t.Setenv to ensure cleanup, although unsetting is the key here if it was set globally.
	// t.Setenv works for setting, unsetting requires manual if we want to be sure?
	// Actually t.Setenv restores the previous value after test.
	// So if we set it to empty string, it unsets it (or sets to empty).
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")

	// Start a mock server that simulates a private internal service
	mockInternalService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Internal Secret Data"))
	}))
	defer mockInternalService.Close()

	hc := &configv1.HttpHealthCheck{
		Url: proto.String(mockInternalService.URL),
		ExpectedCode: proto.Int32(200),
	}

	// This should block since we are using SafeHTTPClient now and we haven't enabled loopback
	err := httpCheckFunc(context.Background(), mockInternalService.URL, hc)

	// Assert that we get an error indicating SSRF block
	assert.Error(t, err, "Should be blocked by SSRF protection")
	if err != nil {
		assert.Contains(t, err.Error(), "ssrf attempt blocked", "Error message should mention ssrf block")
	}
}

func TestHTTPCheckFunc_SSRF_Allowed(t *testing.T) {
	// This test verifies that we CAN bypass the protection if we explicitly allow it via env var.
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Start a mock server
	mockInternalService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockInternalService.Close()

	hc := &configv1.HttpHealthCheck{
		Url: proto.String(mockInternalService.URL),
		ExpectedCode: proto.Int32(200),
	}

	err := httpCheckFunc(context.Background(), mockInternalService.URL, hc)
	assert.NoError(t, err, "Should be allowed when env var is set")
}
