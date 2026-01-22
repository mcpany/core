// Copyright 2026 Author(s) of MCP Any
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
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestUserCRUD(t *testing.T) {
	app := setupTestApp()
	ctx := context.Background()

	// 1. List (Empty)
	t.Run("list empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rr := httptest.NewRecorder()
		app.handleUsers(app.Storage)(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, "[]", rr.Body.String())
	})

	// 2. Create
	var createdID string
	t.Run("create user", func(t *testing.T) {
		user := &configv1.User{
			Id: proto.String("test-user"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						Username:     proto.String("test"),
						PasswordHash: proto.String("plain-password"), // Should be hashed
					},
				},
			},
		}

		// Correctly marshal proto to JSON first
		userBytes, err := protojson.Marshal(user)
		require.NoError(t, err)

		// Wrap in { user: ... }
		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(wrapper)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleUsers(app.Storage)(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp configv1.User
		err = protojson.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "test-user", resp.GetId())
		assert.Equal(t, util.RedactedString, resp.GetAuthentication().GetBasicAuth().GetPasswordHash())

		createdID = resp.GetId()

		// Verify in storage
		stored, err := app.Storage.GetUser(ctx, createdID)
		require.NoError(t, err)
		assert.NotEqual(t, "plain-password", stored.GetAuthentication().GetBasicAuth().GetPasswordHash())
		assert.NotEmpty(t, stored.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	// 3. Get
	t.Run("get user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp configv1.User
		err := protojson.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createdID, resp.GetId())
		assert.Equal(t, util.RedactedString, resp.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	// 4. Update
	t.Run("update user", func(t *testing.T) {
		user := &configv1.User{
			Id: proto.String(createdID),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						Username:     proto.String("test-updated"),
						PasswordHash: proto.String(util.RedactedString), // Keep existing password
					},
				},
			},
		}

		userBytes, err := protojson.Marshal(user)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(wrapper)
		req := httptest.NewRequest(http.MethodPut, "/users/"+createdID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify
		saved, err := app.Storage.GetUser(ctx, createdID)
		require.NoError(t, err)
		assert.Equal(t, "test-updated", saved.GetAuthentication().GetBasicAuth().GetUsername())
		// Password should be preserved (not empty, not REDACTED)
		assert.NotEqual(t, util.RedactedString, saved.GetAuthentication().GetBasicAuth().GetPasswordHash())
		assert.NotEmpty(t, saved.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	// 5. Update Password
	t.Run("update user password", func(t *testing.T) {
		// First get current hash
		savedBefore, _ := app.Storage.GetUser(ctx, createdID)
		hashBefore := savedBefore.GetAuthentication().GetBasicAuth().GetPasswordHash()

		user := &configv1.User{
			Id: proto.String(createdID),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						Username:     proto.String("test-updated"),
						PasswordHash: proto.String("new-password"), // New password
					},
				},
			},
		}

		userBytes, err := protojson.Marshal(user)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(wrapper)
		req := httptest.NewRequest(http.MethodPut, "/users/"+createdID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify
		saved, err := app.Storage.GetUser(ctx, createdID)
		require.NoError(t, err)
		assert.NotEqual(t, hashBefore, saved.GetAuthentication().GetBasicAuth().GetPasswordHash())
		assert.NotEqual(t, "new-password", saved.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	// 6. Delete
	t.Run("delete user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify gone
		saved, err := app.Storage.GetUser(ctx, createdID)
		require.NoError(t, err)
		assert.Nil(t, saved)
	})
}

func TestUserHandlers_Errors(t *testing.T) {
	app := setupTestApp()

	t.Run("Create invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()
		app.handleUsers(app.Storage)(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Create missing id", func(t *testing.T) {
		user := &configv1.User{
			// Missing Id
			Authentication: &configv1.Authentication{},
		}
		body, _ := protojson.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleUsers(app.Storage)(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Get missing id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Get not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/missing", nil)
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Update id mismatch", func(t *testing.T) {
		user := &configv1.User{
			Id: proto.String("other-id"),
		}
		userBytes, _ := protojson.Marshal(user)
		wrapper := map[string]json.RawMessage{"user": json.RawMessage(userBytes)}
		body, _ := json.Marshal(wrapper)
		req := httptest.NewRequest(http.MethodPut, "/users/test-id", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleUserDetail(app.Storage)(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users", nil)
		rr := httptest.NewRecorder()
		app.handleUsers(app.Storage)(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
