// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthTestEndpoint_SSRF(t *testing.T) {
	// Ensure env vars are unset for this test to enforce strict SSRF protection
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "false")

	app := setupTestApp()

	// Create a mock upstream server (on localhost)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer upstream.Close()

	t.Run("should block localhost access by default", func(t *testing.T) {
		reqData := TestAuthRequest{
			TargetURL: upstream.URL, // This is localhost
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		// We expect the request to have failed due to SSRF protection.
		// If the vulnerability exists, resp.Status will be 200 and resp.Body will be "success".
		// If fixed, resp.Error will be present and contain "blocked".

		if resp.Status == 200 && resp.Body == "success" {
			t.Logf("VULNERABILITY CONFIRMED: Successfully accessed localhost: %s", upstream.URL)
			t.Fail() // Fail the test to indicate vulnerability exists (or regression)
		} else {
			assert.NotEmpty(t, resp.Error, "Expected an error message due to blocked connection")
			// The error message comes from util/net.go: "ssrf attempt blocked: ..."
			assert.Contains(t, strings.ToLower(resp.Error), "blocked", "Error message should mention 'blocked'")
		}
	})
}
