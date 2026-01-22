package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRegexValidationRemoteContent(t *testing.T) {
	// Enable loopback secrets for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_SECRETS")

	// Setup a test server that returns "invalid-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid-key"))
	}))
	defer server.Close()

	regex := "^sk-[a-zA-Z0-9]{10}$"
	url := server.URL
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: &url,
			},
		},
		ValidationRegex: &regex,
	}

	// This should fail validation because "invalid-key" does not match regex.
	err := validateSecretValue(context.Background(), secret)

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "secret value does not match validation regex")
	}
}
