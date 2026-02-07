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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestCredentialIDOR(t *testing.T) {
	app := setupTestApp()

	// Setup users
	userA := "userA"
	userB := "userB"

	var createdID string

	// 1. UserA creates a credential
	t.Run("UserA creates credential", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Name: proto.String("UserA Credential"),
			Authentication: configv1.Authentication_builder{
				BearerToken: configv1.BearerTokenAuth_builder{
					Token: configv1.SecretValue_builder{
						PlainText: proto.String("secretA"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(cred)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(body))

		// Inject UserA context
		ctx := auth.ContextWithUser(req.Context(), userA)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		app.createCredentialHandler(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp configv1.Credential
		err := protojson.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		createdID = resp.GetId()
	})

	// 2. UserB lists credentials - should NOT see UserA's credential
	t.Run("UserB lists credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials", nil)

		// Inject UserB context
		ctx := auth.ContextWithUser(req.Context(), userB)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		app.listCredentialsHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var creds []json.RawMessage
		err := json.Unmarshal(rr.Body.Bytes(), &creds)
		require.NoError(t, err)

		found := false
		for _, raw := range creds {
			var c configv1.Credential
			_ = protojson.Unmarshal(raw, &c)
			if c.GetId() == createdID {
				found = true
				break
			}
		}

		if found {
			t.Errorf("IDOR Vulnerability: UserB was able to see UserA's credential %s", createdID)
		}
	})

	// 3. UserB tries to get UserA's credential
	t.Run("UserB gets UserA credential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/"+createdID, nil)

		// Inject UserB context
		ctx := auth.ContextWithUser(req.Context(), userB)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		app.getCredentialHandler(rr, req)

		if rr.Code == http.StatusOK {
			t.Errorf("IDOR Vulnerability: UserB was able to access UserA's credential %s", createdID)
		}
	})

	// 4. UserB tries to delete UserA's credential
	t.Run("UserB deletes UserA credential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/credentials/"+createdID, nil)

		// Inject UserB context
		ctx := auth.ContextWithUser(req.Context(), userB)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		app.deleteCredentialHandler(rr, req)

		if rr.Code == http.StatusNoContent {
			t.Errorf("IDOR Vulnerability: UserB was able to delete UserA's credential %s", createdID)
		}
	})
}
