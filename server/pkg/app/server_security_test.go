// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSecurity_RootEndpoint_PrivilegeEscalation_Repro(t *testing.T) {
	// Setup Application with Auth Manager
	app := NewApplication()

	// Setup Auth Manager with a guest user (no roles)
	password := "guestpassword"
	hashed, _ := passhash.Password(password)

	users := []*configv1.User{
		configv1.User_builder{
			Id: proto.String("guest"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("guest"),
					PasswordHash: proto.String(hashed),
				}.Build(),
			}.Build(),
			Roles: []string{"user"}, // NOT admin
		}.Build(),
	}

	authManager := auth.NewManager()
	authManager.SetUsers(users)
	app.AuthManager = authManager

	// Create Settings Manager (required by middleware)
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil) // No global API key

	// Create Auth Middleware
	authMiddleware := app.createAuthMiddleware(false, false)

	// Create RBAC Middleware (The Fix)
	rbacMiddleware := middleware.NewRBACMiddleware()
	adminOnly := rbacMiddleware.RequireRole("admin")

	// Create a dummy handler representing the Root Handler (JSON-RPC)
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK: Access Granted"))
	})

	// Chain: Auth -> RBAC -> Handler
	// This mirrors the structure in runServerMode for critical endpoints
	protectedHandler := authMiddleware(adminOnly(mockHandler))

	// Create Request
	req := httptest.NewRequest("POST", "/", nil)
	req.SetBasicAuth("guest", password)

	w := httptest.NewRecorder()

	// Execute
	protectedHandler.ServeHTTP(w, req)

	// Verify
	resp := w.Result()

	// ASSERTION: Should be 403 Forbidden because guest does not have "admin" role.
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Guest user should be denied access to root endpoint")

	body := w.Body.String()
	assert.Contains(t, body, "Forbidden", "Response body should indicate forbidden access")
}

func TestSecurity_RootEndpoint_AdminAccess(t *testing.T) {
	// Setup Application with Auth Manager
	app := NewApplication()

	// Setup Auth Manager with an admin user
	password := "adminpassword"
	hashed, _ := passhash.Password(password)

	users := []*configv1.User{
		configv1.User_builder{
			Id: proto.String("admin"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("admin"),
					PasswordHash: proto.String(hashed),
				}.Build(),
			}.Build(),
			Roles: []string{"admin"}, // HAS admin role
		}.Build(),
	}

	authManager := auth.NewManager()
	authManager.SetUsers(users)
	app.AuthManager = authManager

	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	authMiddleware := app.createAuthMiddleware(false, false)
	rbacMiddleware := middleware.NewRBACMiddleware()
	adminOnly := rbacMiddleware.RequireRole("admin")

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK: Access Granted"))
	})

	protectedHandler := authMiddleware(adminOnly(mockHandler))

	req := httptest.NewRequest("POST", "/", nil)
	req.SetBasicAuth("admin", password)

	w := httptest.NewRecorder()
	protectedHandler.ServeHTTP(w, req)

	resp := w.Result()
	// ASSERTION: Should be 200 OK because user has "admin" role.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin user should be granted access")
	assert.Equal(t, "OK: Access Granted", w.Body.String())
}
