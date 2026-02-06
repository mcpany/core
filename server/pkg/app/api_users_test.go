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
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleUsers_List(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Create a user first
	user := configv1.User_builder{Id: proto.String("user1")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	// Inject admin role
	ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []json.RawMessage
	err := json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.Len(t, users, 1)

	var u configv1.User
	err = protojson.Unmarshal(users[0], &u)
	require.NoError(t, err)
	assert.Equal(t, "user1", u.GetId())
}

func TestHandleUserDetail(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserDetail(store)

	// Create a user
	user := configv1.User_builder{Id: proto.String("user1")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	t.Run("Get User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
		// Inject auth context (user accessing self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var u configv1.User
		err := protojson.Unmarshal(w.Body.Bytes(), &u)
		require.NoError(t, err)
		assert.Equal(t, "user1", u.GetId())
	})

	t.Run("Get Non-Existent User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/unknown", nil)
		// Inject admin role to bypass "own user" check and hit 404
		ctx := auth.ContextWithUser(req.Context(), "admin")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update User", func(t *testing.T) {
		updatedUser := configv1.User_builder{
			Id: proto.String("user1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("newpass"),
				}.Build(),
			}.Build(),
		}.Build()
		// Wrap in { user: ... }
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(body))
		// Inject auth context (user accessing self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in store
		u, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		// Password should be hashed (not "newpass")
		assert.NotEqual(t, "newpass", u.GetAuthentication().GetBasicAuth().GetPasswordHash())
		assert.True(t, strings.HasPrefix(u.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
	})

	t.Run("Delete User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/user1", nil)
		// Inject auth context (user deleting self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		u, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Nil(t, u)
	})
}

func TestHashUserPassword_Redaction(t *testing.T) {
	app := NewApplication()
	// Dummy store for testing
	_ = app

	store := memory.NewStore()

	// 1. Create a user with a real hash
	user := configv1.User_builder{
		Id: proto.String("user-redact"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String("real-hash"),
			}.Build(),
		}.Build(),
	}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	// 2. Simulate an update where password_hash is "[REDACTED]"
	updatedUser := configv1.User_builder{
		Id: proto.String("user-redact"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String("REDACTED"),
			}.Build(),
		}.Build(),
	}.Build()

	// 3. Call hashUserPassword
	err := hashUserPassword(context.Background(), updatedUser, store)
	require.NoError(t, err)

	// 4. Verify that the hash was restored to "real-hash"
	assert.Equal(t, "real-hash", updatedUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
}
