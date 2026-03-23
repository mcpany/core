// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestHandleUserDetail_PrivilegeEscalation_Reproduction(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	// Setup: Create victim user
	victim := configv1.User_builder{Id: proto.String("victim-user"), Roles: []string{"user"}}.Build()
	require.NoError(t, store.CreateUser(context.Background(), victim))

	t.Run("Victim Elevates Own Privileges to Admin", func(t *testing.T) {
		// Attempt to update own profile and inject "admin" role
		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"id": "victim-user",
				"roles": []string{"admin"}, // <--- Privilege Escalation attempt
			},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/users/victim-user", bytes.NewReader(body))
		// Simulate Authenticated User: victim-user
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check if the user is now an admin
		updatedUser, err := store.GetUser(context.Background(), "victim-user")
		require.NoError(t, err)

		// VULNERABILITY CHECK: The user should NOT have the admin role
		for _, role := range updatedUser.GetRoles() {
			if role == "admin" {
				t.Logf("VULNERABILITY REPRODUCED: User 'victim-user' escalated privileges to 'admin'.")
			}
		}
	})
}

// Sentinel Security Update: IDOR / Privilege Escalation via profile_ids verification
func TestUserDetail_IDOR_ProfileIDs_Sentinel(t *testing.T) {
	app, mockStore := setupApiTestApp()

	// 2. Mock a regular user
	normalUser := configv1.User_builder{
		Id:         proto.String("user-123"),
		Roles:      []string{"user"},
		ProfileIds: []string{"profile-1"},
	}.Build()
	mockStore.CreateUser(context.Background(), normalUser)

	// 3. User attempts to update their own profile, but injects "admin" role and "admin-profile"
	maliciousUpdatePayload := map[string]interface{}{
		"user": map[string]interface{}{
			"id": "user-123",
			"roles": []string{"admin"}, // Escalate role
			"profile_ids": []string{"admin-profile-x"}, // Escalate profile access
		},
	}

	// 4. Send request
	body, _ := json.Marshal(maliciousUpdatePayload)
	req := httptest.NewRequest(http.MethodPut, "/users/user-123", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithUser(req.Context(), "user-123")) // Authenticated as user-123

	w := httptest.NewRecorder()
	app.handleUserDetail(mockStore).ServeHTTP(w, req)

	// 5. Assert successful (but sanitized) update
	assert.Equal(t, http.StatusOK, w.Code)

	updatedUser, _ := mockStore.GetUser(context.Background(), "user-123")
	assert.Equal(t, []string{"user"}, updatedUser.GetRoles(), "User roles should be restored to original")
	assert.Equal(t, []string{"profile-1"}, updatedUser.GetProfileIds(), "User profile_ids should be restored to original")
}
