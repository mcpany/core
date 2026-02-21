// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_AwsSecretManager(t *testing.T) {
	// Set dummy AWS credentials to prevent SDK from trying to fetch them via IMDS
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	// Allow loopback for testing
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// Create a mock AWS Secrets Manager server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify standard AWS request properties
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify target
		target := r.Header.Get("X-Amz-Target")
		if target != "secretsmanager.GetSecretValue" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"__type": "UnknownOperationException", "Message": "Unknown operation %s"}`, target)
			return
		}

		// Parse request body
		var input struct {
			SecretId string `json:"SecretId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/x-amz-json-1.1")

		// Route based on SecretId
		switch input.SecretId {
		case "my-secret":
			fmt.Fprint(w, `{"SecretString": "my-aws-secret", "Name": "my-secret"}`)
		case "my-secret-binary":
			binaryData := "my-binary-secret"
			encoded := base64.StdEncoding.EncodeToString([]byte(binaryData))
			fmt.Fprintf(w, `{"SecretBinary": "%s", "Name": "my-secret"}`, encoded)
		case "my-secret-json":
			innerJSON := `{\"api_key\": \"12345\", \"db_pass\": \"secret\"}`
			fmt.Fprintf(w, `{"SecretString": "%s", "Name": "my-secret"}`, innerJSON)
		case "my-secret-json-num":
			innerJSON := `{\"port\": 8080}`
			fmt.Fprintf(w, `{"SecretString": "%s", "Name": "my-secret"}`, innerJSON)
		case "my-secret-json-missing":
			innerJSON := `{\"api_key\": \"12345\"}`
			fmt.Fprintf(w, `{"SecretString": "%s", "Name": "my-secret"}`, innerJSON)
		case "my-secret-json-invalid":
			innerJSON := `invalid-json`
			fmt.Fprintf(w, `{"SecretString": "%s", "Name": "my-secret"}`, innerJSON)
		case "my-secret-empty":
			fmt.Fprint(w, `{"Name": "my-secret"}`)
		case "my-secret-error":
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"__type": "ResourceNotFoundException"}`)
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"__type": "ResourceNotFoundException", "Message": "Secret not found"}`)
		}
	}))
	defer ts.Close()

	// Configure AWS SDK to use the mock server
	t.Setenv("AWS_ENDPOINT_URL", ts.URL)

	t.Run("SecretString", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-aws-secret", resolved)
	})

	t.Run("SecretBinary", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-binary")
		smSecret.SetRegion("us-east-1")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-binary-secret", resolved)
	})

	t.Run("SecretString with JSON Key", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-json")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("api_key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "12345", resolved)
	})

	t.Run("SecretString with JSON Key (Non-String Value)", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-json-num")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("port")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "8080", resolved)
	})

	t.Run("SecretString with JSON Key Not Found", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-json-missing")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("missing_key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key \"missing_key\" not found in secret json")
	})

	t.Run("SecretString Invalid JSON", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-json-invalid")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("some_key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal secret json")
	})

	t.Run("No SecretString or SecretBinary", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-empty")
		smSecret.SetRegion("us-east-1")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret value is not a string or binary")
	})

	t.Run("AWS Error", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret-error")
		smSecret.SetRegion("us-east-1")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get secret value from aws secrets manager")
	})
}
