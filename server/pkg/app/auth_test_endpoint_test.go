// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func protoPtr(s string) *string {
	return &s
}

func TestHandleAuthTest_HTTP_Success(t *testing.T) {
	// Allow localhost connections for testing
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// 1. Setup Test Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// 2. Setup Application
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	app.Storage = memory.NewStore()
	handler := app.handleAuthTest()

	// 3. Prepare Request
	reqBody := AuthTestRequest{
		ServiceType: "HTTP",
		ServiceConfig: map[string]any{
			"http_service": map[string]any{
				"address": ts.URL, // Use the test server URL
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(bodyBytes))

	// 4. Execute
	w := httptest.NewRecorder()
	handler(w, req)

	// 5. Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp AuthTestResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Connection verification successful", resp.Message)
}

func TestHandleAuthTest_HTTP_Failure(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// 1. Setup Test Server (Returns 500)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	app := NewApplication()
	app.Storage = memory.NewStore()
	handler := app.handleAuthTest()

	reqBody := AuthTestRequest{
		ServiceType: "HTTP",
		ServiceConfig: map[string]any{
			"http_service": map[string]any{
				"address": ts.URL,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(bodyBytes))

	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Handler returns 200 even on connection check failure
	var resp AuthTestResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "server returned error status: 500")
}

func TestHandleAuthTest_CredentialInjection(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// 1. Setup Test Server checking for API Key
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "secret-123" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	// 2. Setup Application & Credential
	app := NewApplication()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleAuthTest()

	cred := configv1.Credential_builder{
		Id: protoPtr("cred1"),
		Authentication: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				In:        configv1.APIKeyAuth_HEADER.Enum(),
				ParamName: protoPtr("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: protoPtr("secret-123"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()
	require.NoError(t, store.SaveCredential(context.Background(), cred))

	// 3. Request using the credential
	reqBody := AuthTestRequest{
		CredentialID: "cred1",
		ServiceType:  "HTTP",
		ServiceConfig: map[string]any{
			"http_service": map[string]any{
				"address": ts.URL,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(bodyBytes))

	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp AuthTestResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestHandleAuthTest_SSRF_Protection(t *testing.T) {
	// Ensure we DO NOT allow local IPs
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")

	app := NewApplication()
	app.Storage = memory.NewStore()
	handler := app.handleAuthTest()

	// Try to access localhost
	reqBody := AuthTestRequest{
		ServiceType: "HTTP",
		ServiceConfig: map[string]any{
			"http_service": map[string]any{
				"address": "http://127.0.0.1:8080",
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(bodyBytes))

	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp AuthTestResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	// The message should mention unsafe URL or loopback
	assert.True(t, strings.Contains(resp.Message, "unsafe url") || strings.Contains(resp.Message, "loopback"), "Expected security error, got: "+resp.Message)
}

func TestHandleAuthTest_InvalidInput(t *testing.T) {
	app := NewApplication()
	handler := app.handleAuthTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
