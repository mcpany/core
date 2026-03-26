// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRBACContextPopulation(t *testing.T) {
	app := NewApplication()
	app.AuthManager = auth.NewManager()
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	password := "securepassword"
	hash, err := passhash.Password(password)
	assert.NoError(t, err)

	user := configv1.User_builder{
		Id: proto.String("admin-user"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username:     proto.String("admin-user"),
				PasswordHash: proto.String(hash),
			}.Build(),
		}.Build(),
	}.Build()
	user.SetRoles([]string{"admin"})
	app.AuthManager.SetUsers([]*configv1.User{user})

	// Create a middleware
	mw := app.createAuthMiddleware(false, false)

	// Create a handler that checks the context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enforcer := auth.NewRBACEnforcer()
		hasRole := enforcer.HasRoleInContext(r.Context(), "admin")
		if hasRole {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})

	wrappedHandler := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	req.SetBasicAuth("admin-user", password)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected 200 OK, meaning admin role was found in context")
}
