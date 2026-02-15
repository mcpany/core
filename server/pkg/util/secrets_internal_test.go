package util

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// TestSafeSecretClient_SecurityDefaults verifies that the global safeSecretClient
// is configured correctly for security (SSRF protection).
func TestSafeSecretClient_SecurityDefaults(t *testing.T) {
	// Verify Redirects are blocked
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	// CheckRedirect should return an error to stop redirect
	err := safeSecretClient.CheckRedirect(req, []*http.Request{req})
	assert.Error(t, err, "CheckRedirect should return error to block redirects")
	assert.Equal(t, http.ErrUseLastResponse, err, "CheckRedirect should return http.ErrUseLastResponse")

	// Verify SSRF Protection (Loopback blocked by default)
	// We need to ensure the environment variable is NOT set for this test.
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "") // Force disable
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_SECRETS", "") // Force disable

	secret := &configv1.SecretValue{}
	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://127.0.0.1:12345/secret")
	secret.SetRemoteContent(remoteContent)

	_, err = ResolveSecret(context.Background(), secret)
	assert.Error(t, err, "ResolveSecret should fail for loopback IP")
	// The error message from SafeDialer usually contains "blocked" or "private"
	// Adjust expectation based on actual error message if needed, but "blocked" is likely.
	// Common error might be "dial tcp 127.0.0.1:12345: operation not permitted" or custom error.
	// Since we can't see safe_dialer.go easily, we check for generic failure first.
	// If it fails with "connection refused", then SafeDialer is NOT working (it allowed the dial).
	// If it fails with "blocked", it is working.
	// However, if we assume safeSecretClient uses NewSafeDialer, it should block.

	// Note: If SafeDialer is doing its job, we should NOT see "connection refused".
	assert.False(t, strings.Contains(err.Error(), "connection refused"), "Should be blocked by policy, not connection refused")
}

// mockTransport is a simple RoundTripper for testing
type mockTransport struct {
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return nil, errors.New("mock transport not implemented")
}

func TestResolveSecret_NetworkFailures(t *testing.T) {
	// Save original transport
	originalTransport := safeSecretClient.Transport
	defer func() { safeSecretClient.Transport = originalTransport }()

	t.Run("Transport Error", func(t *testing.T) {
		mock := &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("simulated network failure")
			},
		}
		safeSecretClient.Transport = mock

		secret := &configv1.SecretValue{}
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com/secret")
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated network failure")
	})

	t.Run("Timeout Error", func(t *testing.T) {
		mock := &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return nil, context.DeadlineExceeded
			},
		}
		safeSecretClient.Transport = mock

		secret := &configv1.SecretValue{}
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com/secret")
		secret.SetRemoteContent(remoteContent)

		// Use a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := ResolveSecret(ctx, secret)
		assert.Error(t, err)
		// Depending on implementation, it might wrap the error
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "context deadline exceeded"))
	})
}

func TestResolveSecret_ContextCancellation_Internal(t *testing.T) {
	// Verify that context cancellation propagates to the request
	originalTransport := safeSecretClient.Transport
	defer func() { safeSecretClient.Transport = originalTransport }()

	transportCalled := make(chan struct{})
	mock := &mockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			close(transportCalled)
			// Wait for context cancellation
			<-req.Context().Done()
			return nil, req.Context().Err()
		},
	}
	safeSecretClient.Transport = mock

	secret := &configv1.SecretValue{}
	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://example.com/slow")
	secret.SetRemoteContent(remoteContent)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error)
	go func() {
		_, err := ResolveSecret(ctx, secret)
		errChan <- err
	}()

	// Wait for request to reach transport
	select {
	case <-transportCalled:
	case <-time.After(1 * time.Second):
		t.Fatal("Transport was not called")
	}

	cancel()

	select {
	case err := <-errChan:
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceled"))
	case <-time.After(1 * time.Second):
		t.Fatal("ResolveSecret did not return after cancellation")
	}
}
