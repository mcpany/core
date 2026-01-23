package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestCheckConnection_Repro(t *testing.T) {
	// Case 1: Trailing colon "example.com:"
	// Expected: CheckConnection should probably fail gracefully or handle it.
	// Current hypothesis: It calls DialContext with "example.com:", which fails with "unknown port" or similar.
	t.Run("Trailing colon", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		// "example.com:" -> SplitHostPort returns "example.com", "", nil.
		// DialContext("tcp", "example.com:") -> fails or succeeds depending on network.
		// After fix, it should try example.com:80.
		err := util.CheckConnection(ctx, "example.com:")
		if err != nil {
			assert.Contains(t, err.Error(), ":80")
		}
	})

	// Case 2: [::1]: (IPv6 with empty port)
	t.Run("IPv6 Trailing colon", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err := util.CheckConnection(ctx, "[::1]:")
		assert.Error(t, err)
		if err != nil {
			// This might be "ssrf attempt blocked: host ::1 resolved to loopback ip ::1"
			// The original CheckConnection does NOT include port in the SSRF error if resolution fails or is blocked.
			// But DialContext constructs "::1:80" (after fix) or "::1:" (before fix).
			// If it's blocked, it means resolution succeeded.
			// The error message from SafeDialer says "host %s resolved to ..."
			// It doesn't show the full address being dialed.
			// However, if we can trigger a connection refusal (by allowing loopback but picking a closed port), we can see the address.
		}
	})

	// Case 3: Loopback with trailing colon, allowed (to check port)
	t.Run("Loopback Trailing colon Allowed", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		// 127.0.0.1: should default to 127.0.0.1:80
		err := util.CheckConnection(ctx, "127.0.0.1:")
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "127.0.0.1:80")
		}
	})
}
