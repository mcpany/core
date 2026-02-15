package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHandleGetUserPreferences(t *testing.T) {
	// Setup
	store := memory.NewStore()
	app := &Application{Storage: store}

	// Create a user with preferences
	user := configv1.User_builder{
		Id: proto.String("test-user"),
		Preferences: map[string]string{
			"theme": "dark",
		},
	}.Build()
	_ = store.CreateUser(nil, user)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
	// Inject user into context
	ctx := auth.ContextWithUser(req.Context(), "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	app.handleGetUserPreferences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var prefs map[string]string
	_ = json.NewDecoder(w.Body).Decode(&prefs)
	assert.Equal(t, "dark", prefs["theme"])
}

func TestHandleGetUserPreferences_NotFound(t *testing.T) {
	store := memory.NewStore()
	app := &Application{Storage: store}

	req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
	ctx := auth.ContextWithUser(req.Context(), "unknown-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	app.handleGetUserPreferences(w, req)

	// Should return empty preferences, not 404 (as per implementation for handling ephemeral users)
	assert.Equal(t, http.StatusOK, w.Code)
	var prefs map[string]string
	_ = json.NewDecoder(w.Body).Decode(&prefs)
	assert.Empty(t, prefs)
}

func TestHandleUpdateUserPreferences(t *testing.T) {
	store := memory.NewStore()
	app := &Application{Storage: store}

	prefs := map[string]string{
		"dashboard-layout": "{}",
	}
	body, _ := json.Marshal(prefs)
	req := httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewBuffer(body))
	ctx := auth.ContextWithUser(req.Context(), "new-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	app.handleUpdateUserPreferences(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify storage
	user, err := store.GetUser(nil, "new-user")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "{}", user.GetPreferences()["dashboard-layout"])
}
