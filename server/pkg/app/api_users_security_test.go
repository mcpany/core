// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
			assert.Fail(t, "IDOR Vulnerability found!")
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
				t.Fail()
				assert.Fail(t, "Privilege Escalation Vulnerability found!")
			}
		}
	})
}

func TestHashUserPassword_DoS_Prevention(t *testing.T) {
	// Create an extremely long password payload
	longPassword := strings.Repeat("A", 200)

	user := (&configv1.User_builder{
		Id: proto.String("user1"),
		Authentication: (&configv1.Authentication_builder{
			BasicAuth: (&configv1.BasicAuth_builder{
				PasswordHash: proto.String(longPassword),
			}).Build(),
		}).Build(),
	}).Build()

	err := hashUserPassword(context.Background(), user, nil, nil)
	require.Error(t, err)
	// bcrypt rejects > 72 bytes so the error could be from bcrypt or from our custom limit.
	// As long as it errors, the DoS is prevented. We check for either message.
	assert.True(t, strings.Contains(err.Error(), "password exceeds maximum allowed length") || strings.Contains(err.Error(), "bcrypt: password length exceeds 72 bytes"))
}

func TestHandleUserDetail_Put_Payload_Limit(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	// Payload just over 1MB
	largePayload := make([]byte, 1048576+10)
	for i := range largePayload {
		largePayload[i] = 'A'
	}

	req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(largePayload))
	ctx := auth.ContextWithUser(req.Context(), "user1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Since LimitReader is used by MaxBytesReader, the unmarshal/read process will fail with body too large or similar
	assert.Equal(t, http.StatusBadRequest, rr.Code) // MaxBytesReader returns error that translates to Bad Request usually
}
