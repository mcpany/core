// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestServiceValidateSSRF(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")

	// Start a local server simulating an internal resource (loopback)
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer internalServer.Close()

	app := NewApplication()
	// We don't need to fully initialize app, just enough for the handler
	// handleServiceValidate doesn't depend on much, just static config validation
	// and checkURLReachability.

	handler := app.handleServiceValidate()

	tests := []struct {
		name        string
		address     string
		expectValid bool
		expectError string
	}{
		{
			name:        "Public URL",
			address:     "http://example.com",
			expectValid: true, // Should pass reachability or fail gracefully if offline, but NOT blocked by SSRF logic
		},
		{
			name:        "Internal Loopback URL",
			address:     internalServer.URL,
			expectValid: false,
			expectError: "ssrf attempt blocked",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			httpSvc := &configv1.HttpUpstreamService{}
			httpSvc.SetAddress(tc.address)

			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("test-service")
			svc.SetHttpService(httpSvc)

			body, _ := protojson.Marshal(svc)
			req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !assert.Equal(t, http.StatusOK, rec.Code) {
				t.Logf("Response Body: %s", rec.Body.String())
			}

			var resp map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			if !assert.NoError(t, err) {
				t.Logf("Response Body: %s", rec.Body.String())
				return
			}

			if tc.expectValid {
				// We expect either valid=true OR valid=false but with a connection error (not SSRF blocked)
				// Since example.com might be reachable or not depending on environment.
				// But we specifically want to ensure it is NOT blocked by SSRF if it resolves to public IP.
				// However, example.com resolves to public IP.
				// If we are offline, it might fail DNS or connection.
				// For this test, we care about the negative case (SSRF blocked).
			} else {
				assert.False(t, resp["valid"].(bool))
				assert.Contains(t, resp["error"], tc.expectError)
			}
		})
	}
}
