package util_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSecret_AwsSecretManager_Success(t *testing.T) {
	// Allow loopback for mock server
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// Mock AWS Secrets Manager
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		// AWS SDK v2 usually uses POST
		assert.Equal(t, "POST", r.Method)

		// We can check the target header to verify action
		// "X-Amz-Target": "secretsmanager.GetSecretValue"
		target := r.Header.Get("X-Amz-Target")
		assert.Contains(t, target, "GetSecretValue")

		// Return success response
		response := map[string]interface{}{
			"SecretString": "my-aws-secret",
			"Name":         "my-secret-id",
			"ARN":          "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-id-123456",
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Redirect AWS SDK to mock server
	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-secret-id")
	smSecret.SetRegion("us-east-1")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	resolved, err := util.ResolveSecret(context.Background(), secret)
	require.NoError(t, err)
	assert.Equal(t, "my-aws-secret", resolved)
}

func TestResolveSecret_AwsSecretManager_JSONKey(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"SecretString": `{"username":"admin","password":"complex-password"}`,
			"Name":         "my-db-secret",
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-db-secret")
	smSecret.SetRegion("us-east-1")
	smSecret.SetJsonKey("password")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	resolved, err := util.ResolveSecret(context.Background(), secret)
	require.NoError(t, err)
	assert.Equal(t, "complex-password", resolved)
}

func TestResolveSecret_AwsSecretManager_JSONKey_NotFound(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"SecretString": `{"username":"admin"}`,
			"Name":         "my-db-secret",
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-db-secret")
	smSecret.SetRegion("us-east-1")
	smSecret.SetJsonKey("password")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	_, err := util.ResolveSecret(context.Background(), secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key \"password\" not found")
}

func TestResolveSecret_AwsSecretManager_Binary(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Binary secrets are base64 encoded in JSON response usually, but SDK handles decoding?
		// The mock response structure for SDK v2 unmarshal expects "SecretBinary" as bytes (base64 in wire).
		// We send JSON with base64 encoded SecretBinary.
		response := map[string]interface{}{
			"SecretBinary": []byte("my-binary-secret"), // json encoder will base64 this
			"Name":         "my-binary-secret-id",
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-binary-secret-id")
	smSecret.SetRegion("us-east-1")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	resolved, err := util.ResolveSecret(context.Background(), secret)
	require.NoError(t, err)
	assert.Equal(t, "my-binary-secret", resolved)
}
