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
	"github.com/stretchr/testify/mock"
)

// Tests HandleGetUserPreferences and HandleUpdateUserPreferences
func TestUserHandlers(t *testing.T) {
	// Initialize mocks and application
	mockStore := new(MockStore)
	authManager := auth.NewManager()
	app := &Application{
		Storage:     mockStore,
		AuthManager: authManager,
	}

	// Helper to add context with user
	authCtx := func(userID string) context.Context {
		return auth.ContextWithUser(context.Background(), userID)
	}

	t.Run("HandleGetUserPreferences - User in DB", func(t *testing.T) {
		userID := "db-user"
		expectedPrefs := map[string]string{"theme": "dark"}
		user := configv1.User_builder{
			Id:          &userID,
			Preferences: expectedPrefs,
		}.Build()

		mockStore.On("GetUser", mock.Anything, userID).Return(user, nil).Once()

		req, _ := http.NewRequest("GET", "/preferences", nil)
		req = req.WithContext(authCtx(userID))
		rr := httptest.NewRecorder()

		app.HandleGetUserPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var prefs map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &prefs)
		assert.NoError(t, err)
		assert.Equal(t, expectedPrefs, prefs)
	})

	t.Run("HandleGetUserPreferences - User in Memory (Config)", func(t *testing.T) {
		userID := "mem-user"
		expectedPrefs := map[string]string{"theme": "light"}
		user := configv1.User_builder{
			Id:          &userID,
			Preferences: expectedPrefs,
		}.Build()

		// Store returns nil (not found)
		mockStore.On("GetUser", mock.Anything, userID).Return((*configv1.User)(nil), nil).Once()

		// Add to AuthManager
		authManager.SetUsers([]*configv1.User{user})

		req, _ := http.NewRequest("GET", "/preferences", nil)
		req = req.WithContext(authCtx(userID))
		rr := httptest.NewRecorder()

		app.HandleGetUserPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var prefs map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &prefs)
		assert.NoError(t, err)
		assert.Equal(t, expectedPrefs, prefs)
	})

	t.Run("HandleGetUserPreferences - User Not Found", func(t *testing.T) {
		userID := "unknown-user"

		mockStore.On("GetUser", mock.Anything, userID).Return((*configv1.User)(nil), nil).Once()
		// Ensure AuthManager is clean for this user
		authManager.SetUsers([]*configv1.User{})

		req, _ := http.NewRequest("GET", "/preferences", nil)
		req = req.WithContext(authCtx(userID))
		rr := httptest.NewRecorder()

		app.HandleGetUserPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "{}", rr.Body.String())
	})

	t.Run("HandleUpdateUserPreferences - Create Shadow User", func(t *testing.T) {
		userID := "config-user"
		prefs := map[string]string{"dashboard": "grid"}
		body, _ := json.Marshal(prefs)

		// Setup config user in AuthManager
		configUser := configv1.User_builder{
			Id: &userID,
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{Username: &userID}.Build(),
			}.Build(),
		}.Build()
		authManager.SetUsers([]*configv1.User{configUser})

		// Expect GetUser to fail (not in DB)
		mockStore.On("GetUser", mock.Anything, userID).Return((*configv1.User)(nil), nil).Once()

		// Expect CreateUser to be called with copied auth
		mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == userID &&
				u.GetPreferences()["dashboard"] == "grid" &&
				u.GetAuthentication().GetBasicAuth().GetUsername() == userID // Auth copied
		})).Return(nil).Once()

		req, _ := http.NewRequest("POST", "/preferences", bytes.NewBuffer(body))
		req = req.WithContext(authCtx(userID))
		rr := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("HandleUpdateUserPreferences - Update Existing User", func(t *testing.T) {
		userID := "existing-user"
		prefs := map[string]string{"dashboard": "list"}
		body, _ := json.Marshal(prefs)

		existingUser := configv1.User_builder{
			Id: &userID,
		}.Build()

		mockStore.On("GetUser", mock.Anything, userID).Return(existingUser, nil).Once()

		// Expect UpdateUser
		mockStore.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == userID && u.GetPreferences()["dashboard"] == "list"
		})).Return(nil).Once()

		req, _ := http.NewRequest("POST", "/preferences", bytes.NewBuffer(body))
		req = req.WithContext(authCtx(userID))
		rr := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
