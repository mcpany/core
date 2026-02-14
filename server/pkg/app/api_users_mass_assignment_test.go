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

func TestHandleUserDetail_MassAssignment_Reproduction(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	// Setup: Create a regular user
	victimID := "victim-user"
	victim := configv1.User_builder{
		Id:    proto.String(victimID),
		Roles: []string{"user"},
	}.Build()

	require.NoError(t, store.CreateUser(context.Background(), victim))

	t.Run("Victim Escalates Privilege via Mass Assignment", func(t *testing.T) {
		// Payload attempts to add "admin" role
		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"id":    victimID,
				"roles": []string{"admin"},
			},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/users/"+victimID, bytes.NewReader(body))
		// Simulate Authenticated User: victim-user
		ctx := auth.ContextWithUser(req.Context(), victimID)
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// If successful, it returns 200 OK
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify if the role was updated in the store
		updatedUser, err := store.GetUser(context.Background(), victimID)
		require.NoError(t, err)

		// Correct Behavior Check: Ensure "admin" role is NOT present
		for _, role := range updatedUser.GetRoles() {
			if role == "admin" {
				t.Fatalf("Security Violation: User 'victim-user' escalated to 'admin'.")
			}
		}
	})
}
