// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUserDetail_IDOR_Reproduction(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	// Setup: Create 2 users
	// Uses Builder pattern to support opaque API
	victim := configv1.User_builder{Id: proto.String("victim-user"), Roles: []string{"user"}}.Build()
	admin := configv1.User_builder{Id: proto.String("admin-user"), Roles: []string{"admin"}}.Build()

	require.NoError(t, store.CreateUser(context.Background(), victim))
	require.NoError(t, store.CreateUser(context.Background(), admin))

	t.Run("Victim Access Own Profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/victim-user", nil)
		// Simulate Authenticated User: victim-user
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Victim Access Admin Profile (IDOR)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/admin-user", nil)
		// Simulate Authenticated User: victim-user
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// VULNERABILITY CHECK: Currently this likely returns 200 OK
		if w.Code == http.StatusOK {
			t.Logf("VULNERABILITY REPRODUCED: User 'victim-user' accessed 'admin-user' profile.")
			t.Fail()
		} else {
			assert.Equal(t, http.StatusForbidden, w.Code)
		}
	})

	t.Run("Admin Access Victim Profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/victim-user", nil)
		// Simulate Authenticated User: admin-user
		ctx := auth.ContextWithUser(req.Context(), "admin-user")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleUserDetail_PrivilegeEscalation(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	// Setup: Create a regular user
	attacker := configv1.User_builder{Id: proto.String("attacker"), Roles: []string{"user"}}.Build()
	require.NoError(t, store.CreateUser(context.Background(), attacker))

	t.Run("Attacker Elevates to Admin", func(t *testing.T) {
		// Payload trying to add "admin" role
		payload := `{"user": {"id": "attacker", "roles": ["admin"]}}`
		req := httptest.NewRequest(http.MethodPut, "/users/attacker", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")

		// Simulate Authenticated User: attacker
		ctx := auth.ContextWithUser(req.Context(), "attacker")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// We verify the state in the store
		updatedUser, err := store.GetUser(context.Background(), "attacker")
		require.NoError(t, err)

		hasAdmin := false
		for _, r := range updatedUser.GetRoles() {
			if r == "admin" {
				hasAdmin = true
				break
			}
		}

		if hasAdmin {
			t.Logf("VULNERABILITY REPRODUCED: User 'attacker' successfully added 'admin' role.")
			t.Fail()
		} else {
			// Fix implementation: We now silently ignore the change, so we expect 200 OK
			// but the role should NOT be admin.
			assert.Equal(t, http.StatusOK, w.Code)
			// Ensure roles are still correct (i.e. kept original)
			// attacker started with ["user"]
			assert.Equal(t, []string{"user"}, updatedUser.GetRoles())
		}
	})
}
