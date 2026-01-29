package util //nolint:revive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSecret_ContextCancellation(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
	// Start a test server that hangs to simulate slow response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // Wait for client to cancel
	}))
	defer ts.Close()

	secret := &configv1.SecretValue{}
	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl(ts.URL)
	secret.SetRemoteContent(remoteContent)

	t.Run("Context Cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := ResolveSecret(ctx, secret)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Context Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := ResolveSecret(ctx, secret)
		require.Error(t, err)
		// Error message might vary depending on whether it's "context deadline exceeded" or wrapped
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
