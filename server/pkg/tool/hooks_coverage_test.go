package tool

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewWebhookHook(t *testing.T) {
	t.Parallel()
	config := configv1.WebhookConfig_builder{
		Url:           "http://example.com",
		WebhookSecret: "dGVzdC1zZWNyZXQtdmFsaWQtYmFzZTY0",
	}.Build()
	hook := NewWebhookHook(config)
	assert.NotNil(t, hook)
	assert.Equal(t, "http://example.com", hook.client.url)
	assert.NotNil(t, hook.client.webhook)
}

func TestSigningRoundTripper_RoundTrip(t *testing.T) {
	t.Parallel()
	// "secret" in base64 is c2VjcmV0, but specific library might require more.
	// "test-secret-valid-base64" -> "dGVzdC1zZWNyZXQtdmFsaWQtYmFzZTY0"
	validSecret := "test-secret-valid-base64"
	encodedSecret := base64.StdEncoding.EncodeToString([]byte(validSecret))
	config := configv1.WebhookConfig_builder{
		Url:           "http://example.com",
		WebhookSecret: encodedSecret,
	}.Build()
	hook := NewWebhookHook(config)

	// Create a round tripper independently to test it
	rt := hook.client.client.Transport.(*SigningRoundTripper)
	require.NotNil(t, rt)

	// Create a request
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("test body"))
	require.NoError(t, err)

	// Mock base round tripper
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify signature headers are present
			assert.NotEmpty(t, req.Header.Get("webhook-signature"))
			assert.NotEmpty(t, req.Header.Get("webhook-id"))
			assert.NotEmpty(t, req.Header.Get("webhook-timestamp"))
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil
		},
	}
	rt.base = mockTransport

	resp, err := rt.RoundTrip(req)
	if err == nil {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()
	}
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

type MockTransport struct {
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil
}

func TestWebhookHook_ExecutePre(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Respond with CloudEvent in Binary Mode
		w.Header().Set("Ce-Id", "resp-id")
		w.Header().Set("Ce-Specversion", "1.0")
		w.Header().Set("Ce-Type", "resp-type")
		w.Header().Set("Ce-Source", "server")
		w.Header().Set("Content-Type", "application/json")

		data := map[string]any{
			"allowed": true,
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	config := configv1.WebhookConfig_builder{
		Url: server.URL,
	}.Build()
	hook := NewWebhookHook(config)

	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{}`),
	}

	action, _, err := hook.ExecutePre(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, ActionAllow, action)
}

func TestWebhookHook_ExecutePre_Errors(t *testing.T) {
	t.Parallel()
	// 1. Unmarshal Inputs Fail
	hook := NewWebhookHook(configv1.WebhookConfig_builder{}.Build())
	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{invalid`),
	}
	action, _, err := hook.ExecutePre(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ActionDeny, action)
	assert.Contains(t, err.Error(), "failed to unmarshal inputs")

	// 2. Webhook Call Fail (Network)
	hook.client.url = "http://invalid-url-that-fails-dns-lookup"
	req.ToolInputs = []byte(`{}`)
	action, _, err = hook.ExecutePre(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ActionDeny, action)
	assert.Contains(t, err.Error(), "webhook error")

	// 3. Webhook 500 Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()
	hook.client.url = server.URL
	action, _, err = hook.ExecutePre(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ActionDeny, action)
}

func TestPolicyHook_InvalidRegex(t *testing.T) {
	t.Parallel()
	// Case: Invalid Regex (Should skip and log error, not panic)
	// We verify it falls back to default action (Deny in this case)
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_DENY.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				NameRegex: proto.String("["),
				Action:    configv1.CallPolicy_ALLOW.Enum(),
			}.Build(),
		},
	}.Build()
	hook := NewPolicyHook(policy)
	req := &ExecutionRequest{ToolName: "any", ToolInputs: []byte(`{}`)}
	action, _, err := hook.ExecutePre(context.Background(), req)
	assert.Error(t, err) // Denied by default
	assert.Equal(t, ActionDeny, action)
	assert.Contains(t, err.Error(), "denied by default policy")
}
