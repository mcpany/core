// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOAuth2Auth_Authenticate_Discovery(t *testing.T) {
	// Create a mock OIDC provider server
	// We use a variable to break the circular dependency in the closure
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":         serverURL, // Use serverURL which includes http://ip:port
				"token_endpoint": serverURL + "/token",
			})
			return
		}
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "discovered-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	serverURL = server.URL

	ctx := context.Background()
	clientID := configv1.SecretValue_builder{
		PlainText: proto.String("id"),
	}.Build()
	clientSecret := configv1.SecretValue_builder{
		PlainText: proto.String("secret"),
	}.Build()

	// Create authenticator with IssuerURL only
	authConfig := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId:     clientID,
			ClientSecret: clientSecret,
			IssuerUrl:    proto.String(server.URL),
			// TokenUrl is intentionally missing
		}.Build(),
	}.Build()

	auth, err := NewUpstreamAuthenticator(authConfig)
	require.NoError(t, err)
	require.NotNil(t, auth)

	// Cast to OAuth2Auth to verify internal state before discovery
	oauthAuth, ok := auth.(*OAuth2Auth)
	require.True(t, ok)
	assert.Equal(t, "", oauthAuth.TokenURL)
	assert.Equal(t, server.URL, oauthAuth.IssuerURL)

	// Perform authentication (should trigger discovery)
	req, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
	err = auth.Authenticate(req)
	require.NoError(t, err)

	// Verify header is set
	assert.Equal(t, "Bearer discovered-token", req.Header.Get("Authorization"))

	// Verify TokenURL was populated
	assert.Equal(t, server.URL+"/token", oauthAuth.TokenURL)
}

func TestOAuth2Auth_Authenticate_Discovery_Fail(t *testing.T) {
	// Mock server that returns 404 for discovery
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	ctx := context.Background()
	clientID := configv1.SecretValue_builder{
		PlainText: proto.String("id"),
	}.Build()
	clientSecret := configv1.SecretValue_builder{
		PlainText: proto.String("secret"),
	}.Build()

	authConfig := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId:     clientID,
			ClientSecret: clientSecret,
			IssuerUrl:    proto.String(server.URL),
		}.Build(),
	}.Build()

	auth, err := NewUpstreamAuthenticator(authConfig)
	require.NoError(t, err)

	req, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
	err = auth.Authenticate(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover OIDC configuration")
}
