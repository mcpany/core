package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestSecretCRUD(t *testing.T) {
	app := NewApplication()
	app.Storage = memory.NewStore()
	ctx := context.Background()

	// 1. Create Secret
	var secretID string
	t.Run("Create Secret", func(t *testing.T) {
		secret := configv1.Secret_builder{
			Name:  proto.String("OpenAI Key"),
			Key:   proto.String("OPENAI_API_KEY"),
			Value: proto.String("sk-12345"),
		}.Build()
		body, _ := protojson.Marshal(secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.createSecretHandler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp configv1.Secret
		err := protojson.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.GetId())
		secretID = resp.GetId()
		assert.Equal(t, "OpenAI Key", resp.GetName())
		// Check masking
		assert.Equal(t, "********", resp.GetValue())
	})

	// 2. List Secrets
	t.Run("List Secrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
		w := httptest.NewRecorder()
		app.listSecretsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Assuming listSecretsHandler returns { "secrets": [...] }
		var wrapper struct {
			Secrets []json.RawMessage `json:"secrets"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &wrapper)
		require.NoError(t, err)
		assert.Len(t, wrapper.Secrets, 1)

		var s configv1.Secret
		err = protojson.Unmarshal(wrapper.Secrets[0], &s)
		require.NoError(t, err)
		assert.Equal(t, secretID, s.GetId())
		assert.Equal(t, "********", s.GetValue())
	})

	// 3. Get Secret
	t.Run("Get Secret", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/"+secretID, nil)
		w := httptest.NewRecorder()
		app.getSecretHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var s configv1.Secret
		err := protojson.Unmarshal(w.Body.Bytes(), &s)
		require.NoError(t, err)
		assert.Equal(t, secretID, s.GetId())
		assert.Equal(t, "********", s.GetValue())
	})

	// 4. Reveal Secret
	t.Run("Reveal Secret", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets/"+secretID+"/reveal", nil)
		w := httptest.NewRecorder()
		app.revealSecretHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "sk-12345", resp["value"])
	})

	// 5. Delete Secret
	t.Run("Delete Secret", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/"+secretID, nil)
		w := httptest.NewRecorder()
		app.deleteSecretHandler(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify gone
		s, err := app.Storage.GetSecret(ctx, secretID)
		require.NoError(t, err)
		assert.Nil(t, s)
	})
}
