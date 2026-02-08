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
			expectValid: true, // Validation no longer probes, so it doesn't trigger SSRF check
			expectError: "",
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
				// Static validation only, so it should be valid
				assert.True(t, resp["valid"].(bool))
			} else {
				assert.False(t, resp["valid"].(bool))
				assert.Contains(t, resp["error"], tc.expectError)
			}
		})
	}
}
