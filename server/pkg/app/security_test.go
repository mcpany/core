package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util/passhash"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSecurity_RootAccess_Enforcement(t *testing.T) {
	// Setup
	app := NewApplication()
	authManager := auth.NewManager()

	// Create a hash for "password"
	hash, err := passhash.Password("password")
	require.NoError(t, err)

	// Create a non-admin user
	guestUser := configv1.User_builder{
		Id: proto.String("guest"),
		Roles: []string{"guest"},
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("guest"),
				PasswordHash: proto.String(hash),
			}.Build(),
		}.Build(),
	}.Build()

	// Create an admin user
	adminUser := configv1.User_builder{
		Id: proto.String("admin"),
		Roles: []string{"admin"},
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("admin"),
				PasswordHash: proto.String(hash),
			}.Build(),
		}.Build(),
	}.Build()

	authManager.SetUsers([]*configv1.User{guestUser, adminUser})
	app.AuthManager = authManager

    // Set up SettingsManager
    app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// Middleware
	middleware := app.createAuthMiddleware(false, false)

	// Handler simulating root access
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Access Granted"))
	}))

	t.Run("Guest_Blocked", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.SetBasicAuth("guest", "password")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, "Security Check: Guest should be blocked (403)")
	})

	t.Run("Admin_Allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.SetBasicAuth("admin", "password")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Admin should be allowed")
	})
}
