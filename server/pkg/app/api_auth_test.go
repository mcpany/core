// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestHandleInitiateOAuth_MethodNotAllowed(t *testing.T) {
	app := NewApplication()
	handler := http.HandlerFunc(app.handleInitiateOAuth)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/initiate", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleOAuthCallback(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	app.AuthManager.SetStorage(store)
	handler := http.HandlerFunc(app.handleOAuthCallback)

	// Create a service with OAuth2 config
	clientId := &configv1.SecretValue{}
	clientId.SetPlainText("client-id")
	clientSecret := &configv1.SecretValue{}
	clientSecret.SetPlainText("client-secret")

	oauth2Auth := &configv1.OAuth2Auth{}
	oauth2Auth.SetClientId(clientId)
	oauth2Auth.SetClientSecret(clientSecret)
	oauth2Auth.SetTokenUrl("https://example.com/token")
	oauth2Auth.SetScopes("read write")

	authMethod := &configv1.Authentication{}
	authMethod.SetOauth2(oauth2Auth)

	service := &configv1.UpstreamServiceConfig{}
	service.SetName("test-service")
	service.SetId("test-service")
	service.SetUpstreamAuth(authMethod)
	require.NoError(t, store.SaveService(context.Background(), service))

	t.Run("Success", func(t *testing.T) {
		reqBody := map[string]string{
			"service_id":   "test-service",
			"code":         "auth-code",
			"redirect_url": "http://localhost:3000/callback",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/callback", bytes.NewReader(body))
		req = req.WithContext(auth.ContextWithUser(req.Context(), "user1"))

		// Mock HTTP Client for Token Exchange
		mockClient := &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request
					if req.URL.String() == "https://example.com/token" {
						respBody := map[string]interface{}{
							"access_token":  "access-token-123",
							"token_type":    "Bearer",
							"refresh_token": "refresh-token-123",
							"expiry":        time.Now().Add(1 * time.Hour).Format(time.RFC3339),
							"scope":         "read write",
						}
						b, _ := json.Marshal(respBody)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(b)),
							Header:     make(http.Header),
						}, nil
					}
					return &http.Response{StatusCode: http.StatusNotFound}, nil
				},
			},
		}
		// Inject mock client into context
		ctx := context.WithValue(req.Context(), oauth2.HTTPClient, mockClient)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")

		// Verify token saved
		token, err := store.GetToken(context.Background(), "user1", "test-service")
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "access-token-123", token.GetAccessToken())
	})

	t.Run("Exchange Error", func(t *testing.T) {
		reqBody := map[string]string{
			"service_id":   "test-service",
			"code":         "bad-code",
			"redirect_url": "http://localhost:3000/callback",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/callback", bytes.NewReader(body))
		req = req.WithContext(auth.ContextWithUser(req.Context(), "user1"))

		// Mock HTTP Client for Token Exchange failure
		mockClient := &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader(`{"error": "invalid_grant"}`)),
						Header:     make(http.Header),
					}, nil
				},
			},
		}
		ctx := context.WithValue(req.Context(), oauth2.HTTPClient, mockClient)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to handle callback")
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/oauth/callback", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
