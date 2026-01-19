package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleCredentials_Security_Redaction(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()
	app.Storage = store

	// Create credential with sensitive data
	cred := &configv1.Credential{
		Id:   proto.String("cred1"),
		Name: proto.String("Test Cred"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BearerToken{
				BearerToken: &configv1.BearerTokenAuth{
					Token: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{
							PlainText: "my-secret-token",
						},
					},
				},
			},
		},
		Token: &configv1.UserToken{
			AccessToken:  proto.String("access-token-123"),
			RefreshToken: proto.String("refresh-token-456"),
		},
	}
	require.NoError(t, store.SaveCredential(context.Background(), cred))

	t.Run("ListCredentials_ShouldNotLeakSecrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		assert.NotContains(t, body, "my-secret-token", "Bearer token should be redacted")
		assert.NotContains(t, body, "access-token-123", "Access token should be redacted")
		assert.NotContains(t, body, "refresh-token-456", "Refresh token should be redacted")
	})
}
