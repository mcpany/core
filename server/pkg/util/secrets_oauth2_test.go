package util_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret_RemoteContent_OAuth2(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// 1. Mock Token Endpoint
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		clientID := r.FormValue("client_id")
		clientSecret := r.FormValue("client_secret")
		grantType := r.FormValue("grant_type")

		if clientID == "" {
			user, pass, ok := r.BasicAuth()
			if ok {
				clientID = user
				clientSecret = pass
			}
		}

		if clientID != "test-client-id" || clientSecret != "test-client-secret" || grantType != "client_credentials" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenServer.Close()

	// 2. Mock Secret Endpoint
	secretServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer mock-access-token" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.Write([]byte("secret-value"))
	}))
	defer secretServer.Close()

	// 3. Configure SecretValue
	secret := configv1.SecretValue_builder{
		RemoteContent: configv1.RemoteContent_builder{
			HttpUrl: proto.String(secretServer.URL),
			Auth: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					TokenUrl: proto.String(tokenServer.URL),
					ClientId: configv1.SecretValue_builder{
						PlainText: proto.String("test-client-id"),
					}.Build(),
					ClientSecret: configv1.SecretValue_builder{
						PlainText: proto.String("test-client-secret"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	// 4. Resolve Secret
	val, err := util.ResolveSecret(context.Background(), secret)

	// Expectation: This should succeed now (after fix)
	assert.NoError(t, err)
	assert.Equal(t, "secret-value", val)
}
