package tool

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webhook "github.com/standard-webhooks/standard-webhooks/libraries/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigningRoundTripper(t *testing.T) {
	// Create a real webhook signer
	// Secret must be base64
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	wh, err := webhook.NewWebhook(secret)
	require.NoError(t, err)

	// Mock the base transport
	mockTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify headers are present
			assert.NotEmpty(t, req.Header.Get("Webhook-Id"))
			assert.NotEmpty(t, req.Header.Get("Webhook-Timestamp"))
			assert.NotEmpty(t, req.Header.Get("Webhook-Signature"))

			// Verify body is intact
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body.Close()

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)), // Echo body
			}, nil
		},
	}

	rt := &SigningRoundTripper{
		signer: wh,
		base:   mockTransport,
	}

	t.Run("With ContentLength", func(t *testing.T) {
		payload := "test payload"
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(payload))
		req.ContentLength = int64(len(payload)) // Explicitly set

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, payload, string(body))
	})

	t.Run("Without ContentLength", func(t *testing.T) {
		payload := "test payload no len"
		// buffer doesn't set ContentLength on request automatically unless NewRequest does?
		// NewRequest usually sets it.
		// We force it to -1
		req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(payload))
		req.ContentLength = -1

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, payload, string(body))
	})
}

type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}
