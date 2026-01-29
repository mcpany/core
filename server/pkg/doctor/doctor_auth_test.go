package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRunChecks_Authentication_ActiveVerification(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Start a mock HTTP server that enforces authentication
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API Key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "secret123" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check Bearer Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer valid-token" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check Basic Auth
		user, pass, ok := r.BasicAuth()
		if ok && user == "user" && pass == "pass" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("api-key-success"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						ParamName: strPtr("X-API-Key"),
						Value: configv1.SecretValue_builder{
							PlainText: proto.String("secret123"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("api-key-fail"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						ParamName: strPtr("X-API-Key"),
						Value: configv1.SecretValue_builder{
							PlainText: proto.String("wrong-key"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("bearer-success"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							PlainText: proto.String("valid-token"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("bearer-fail"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							PlainText: proto.String("invalid-token"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("basic-success"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BasicAuth: configv1.BasicAuth_builder{
						Username: strPtr("user"),
						Password: configv1.SecretValue_builder{
							PlainText: proto.String("pass"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("basic-fail"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					BasicAuth: configv1.BasicAuth_builder{
						Username: strPtr("user"),
						Password: configv1.SecretValue_builder{
							PlainText: proto.String("wrongpass"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("no-auth-401"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
				// No Auth configured
			}.Build(),
		},
	}.Build()

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 7)

	// API Key
	assert.Equal(t, "api-key-success", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, "api-key-fail", results[1].ServiceName)
	assert.Equal(t, StatusError, results[1].Status)
	assert.Contains(t, results[1].Message, "Authentication failed (401 Unauthorized)")

	// Bearer Token
	assert.Equal(t, "bearer-success", results[2].ServiceName)
	assert.Equal(t, StatusOk, results[2].Status)
	assert.Equal(t, "bearer-fail", results[3].ServiceName)
	assert.Equal(t, StatusError, results[3].Status)

	// Basic Auth
	assert.Equal(t, "basic-success", results[4].ServiceName)
	assert.Equal(t, StatusOk, results[4].Status)
	assert.Equal(t, "basic-fail", results[5].ServiceName)
	assert.Equal(t, StatusError, results[5].Status)

	// No Auth
	assert.Equal(t, "no-auth-401", results[6].ServiceName)
	assert.Equal(t, StatusWarning, results[6].Status)
	assert.Contains(t, results[6].Message, "Service reachable but returned: 401 Unauthorized")
}
