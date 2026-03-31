// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSSOMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	t.Run("Disabled", func(t *testing.T) {
		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(false),
		}.Build()
		mw := SSOMiddleware(cfg)
		handler := mw(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("Enabled_MissingAuth", func(t *testing.T) {
		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(true),
			IdpUrl:  proto.String("https://idp.example.com"),
		}.Build()
		mw := SSOMiddleware(cfg)
		handler := mw(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Authentication required", resp["error"])
		assert.Equal(t, "https://idp.example.com/login", resp["login_url"])
	})

	t.Run("Enabled_InvalidBearerToken", func(t *testing.T) {
		// Mock IDP server returning 401 for invalid tokens
		mockIDP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/userinfo", r.URL.Path)
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer mockIDP.Close()

		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(true),
			IdpUrl:  proto.String(mockIDP.URL),
		}.Build()
		mw := SSOMiddleware(cfg)
		handler := mw(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Enabled_ValidBearerToken", func(t *testing.T) {
		// Mock IDP server returning 200 with userinfo payload
		mockIDP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/userinfo", r.URL.Path)
			assert.Equal(t, "Bearer valid-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"sub": "auth-user-123", "email": "test@example.com"}`))
		}))
		defer mockIDP.Close()

		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(true),
			IdpUrl:  proto.String(mockIDP.URL),
		}.Build()
		mw := SSOMiddleware(cfg)

		var capturedUserID string
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "auth-user-123", capturedUserID)
	})

	t.Run("Enabled_ValidBearerToken_EmailFallback", func(t *testing.T) {
		// Mock IDP server returning 200 without sub but with email
		mockIDP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/userinfo", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"email": "fallback@example.com"}`))
		}))
		defer mockIDP.Close()

		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(true),
			IdpUrl:  proto.String(mockIDP.URL),
		}.Build()
		mw := SSOMiddleware(cfg)

		var capturedUserID string
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer token-no-sub")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "fallback@example.com", capturedUserID)
	})

	t.Run("Enabled_TrustedProxyHeader", func(t *testing.T) {
		cfg := configv1.SSOConfig_builder{
			Enabled: proto.Bool(true),
			IdpUrl:  proto.String("https://idp.example.com"),
		}.Build()
		mw := SSOMiddleware(cfg)
		var capturedUserID string
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-MCP-Identity", "alice")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "alice", capturedUserID)
	})
}
