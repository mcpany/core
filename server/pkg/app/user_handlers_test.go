package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestHandleGetUserPreferences(t *testing.T) {
	app := NewApplication()
	mockStore := new(MockStore)
	app.Storage = mockStore

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("User Found", func(t *testing.T) {
		userID := "test-user"
		user := configv1.User_builder{
			Id: proto.String(userID),
			Preferences: map[string]string{
				"theme": "dark",
			},
		}.Build()
		mockStore.On("GetUser", mock.Anything, userID).Return(user, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		json.Unmarshal(w.Body.Bytes(), &prefs)
		assert.Equal(t, "dark", prefs["theme"])
	})

	t.Run("User Not Found", func(t *testing.T) {
		userID := "unknown-user"
		mockStore.On("GetUser", mock.Anything, userID).Return((*configv1.User)(nil), assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		json.Unmarshal(w.Body.Bytes(), &prefs)
		assert.Empty(t, prefs)
	})
}

func TestHandleUpdateUserPreferences(t *testing.T) {
	app := NewApplication()
	mockStore := new(MockStore)
	app.Storage = mockStore
	authManager := auth.NewManager()
	app.AuthManager = authManager

	t.Run("Update Existing User", func(t *testing.T) {
		userID := "test-user"
		user := configv1.User_builder{
			Id: proto.String(userID),
			Preferences: map[string]string{
				"theme": "light",
			},
		}.Build()
		newPrefs := map[string]string{"theme": "dark"}
		body, _ := json.Marshal(newPrefs)

		mockStore.On("GetUser", mock.Anything, userID).Return(user, nil).Once()
		mockStore.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == userID && u.GetPreferences()["theme"] == "dark"
		})).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create New User", func(t *testing.T) {
		userID := "new-user"
		newPrefs := map[string]string{"theme": "dark"}
		body, _ := json.Marshal(newPrefs)

		// Mock config user for copy check
		configUser := configv1.User_builder{
			Id:    proto.String(userID),
			Roles: []string{"admin"},
		}.Build()
		authManager.SetUsers([]*configv1.User{configUser})

		mockStore.On("GetUser", mock.Anything, userID).Return((*configv1.User)(nil), assert.AnError).Once()
		mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			// Check if roles were copied
			hasRole := false
			for _, r := range u.GetRoles() {
				if r == "admin" {
					hasRole = true
				}
			}
			return u.GetId() == userID && u.GetPreferences()["theme"] == "dark" && hasRole
		})).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
