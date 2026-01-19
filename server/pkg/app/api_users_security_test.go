package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUsers_Security_Redaction(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.configPaths = []string{}
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Create user with sensitive data
	loc := configv1.APIKeyAuth_HEADER
	user := &configv1.User{
		Id: proto.String("secure-user"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					VerificationValue: proto.String("super-secret-key"),
					ParamName:         proto.String("X-API-Key"),
					In:                &loc,
				},
			},
		},
	}
	require.NoError(t, store.CreateUser(context.Background(), user))

	t.Run("ListUsers_ShouldNotLeakSecrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// This should now PASS
		assert.NotContains(t, body, "super-secret-key", "VerificationValue should be redacted")
	})
}
