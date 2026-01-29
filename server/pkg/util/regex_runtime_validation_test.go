package util_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret_ValidationRegex(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	t.Run("PlainText with valid regex", func(t *testing.T) {
		secret := &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{
				PlainText: "valid-key-123",
			},
			ValidationRegex: proto.String("^[a-z]+-[a-z]+-[0-9]+$"),
		}
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "valid-key-123", resolved)
	})

	t.Run("PlainText with invalid regex", func(t *testing.T) {
		secret := &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{
				PlainText: "invalid-key",
			},
			ValidationRegex: proto.String("^[0-9]+$"), // Expects only numbers
		}
		_, err := util.ResolveSecret(context.Background(), secret)

		// This should fail, but currently passes (bug)
		assert.Error(t, err, "Expected error because secret does not match regex")
		if err != nil {
			assert.Contains(t, err.Error(), "secret value does not match validation regex")
		}
	})

	t.Run("RemoteContent with invalid regex", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("remote-secret-value"))
		}))
		defer server.Close()

		secret := &configv1.SecretValue{
			Value: &configv1.SecretValue_RemoteContent{
				RemoteContent: &configv1.RemoteContent{
					HttpUrl: proto.String(server.URL),
				},
			},
			ValidationRegex: proto.String("^sk-[a-z]+$"), // Expects sk- prefix
		}

		_, err := util.ResolveSecret(context.Background(), secret)

		// This should fail, but currently passes (bug)
		assert.Error(t, err, "Expected error because remote secret does not match regex")
		if err != nil {
			assert.Contains(t, err.Error(), "secret value does not match validation regex")
		}
	})
}
