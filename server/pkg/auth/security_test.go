package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSecurity_BypassProtection(t *testing.T) {
	// Setup Manager
	manager := NewManager()
	password := "correct-password"
	hash, _ := passhash.Password(password)

	users := []*configv1.User{
		// User with Basic Auth
		configv1.User_builder{
			Id: proto.String("valid-user"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String(hash),
				}.Build(),
			}.Build(),
		}.Build(),
		// User WITHOUT Basic Auth (e.g. OIDC only)
		configv1.User_builder{
			Id: proto.String("oidc-user"),
			Authentication: configv1.Authentication_builder{
				Oidc: configv1.OIDCAuth_builder{
					Issuer: proto.String("https://accounts.google.com"),
				}.Build(),
			}.Build(),
		}.Build(),
	}
	manager.SetUsers(users)

	t.Run("Valid user, correct password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetBasicAuth("valid-user", password)
		_, err := manager.Authenticate(context.Background(), "", req)
		assert.NoError(t, err)
	})

	t.Run("Valid user, wrong password", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("valid-user", "wrong-password")
		_, err := manager.Authenticate(context.Background(), "", req)
		assert.Error(t, err)
	})

	t.Run("Invalid user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("invalid-user", password)
		_, err := manager.Authenticate(context.Background(), "", req)
		assert.Error(t, err)
	})

	t.Run("User without Basic Auth (Bypass Attempt)", func(t *testing.T) {
		// Attempt to login as 'oidc-user' using the dummy password string.
		// If the system checks against dummyHash and allows login, this is a vulnerability.
		// Note: We don't expose the dummy password string publicly, but for testing we assume
		// the attacker guesses it or it's leaked.
		dummyPassword := "dummy-password-for-timing-mitigation"

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("oidc-user", dummyPassword)
		_, err := manager.Authenticate(context.Background(), "", req)
		assert.Error(t, err, "Should fail even if password matches dummy hash")
		// The error returned is the final fallback error from Manager.Authenticate
		// "unauthorized: no authentication configured"
	})
}
