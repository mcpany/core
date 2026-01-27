// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_AwsSecretManager_Binary_Bug(t *testing.T) {
	// Mock AWS Secrets Manager Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		if target == "secretsmanager.GetSecretValue" {
			var req struct {
				SecretId string
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			w.WriteHeader(http.StatusOK)

			if req.SecretId == "my-binary-raw" {
				// Binary with trailing newline
				secretContent := []byte("raw-value\n")
				resp := map[string]interface{}{
					"Name":         "my-binary-raw",
					"SecretBinary": secretContent,
				}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}

			// Default: Secret content is JSON, but stored as binary
			secretContent := []byte(`{"bin-key": "bin-value"}`)
			resp := map[string]interface{}{
				"Name":         "my-binary-secret",
				"SecretBinary": secretContent,
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	t.Run("AwsSecretManager Binary with Key", func(t *testing.T) {
		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-binary-secret")
		awsSecret.SetJsonKey("bin-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)

		// BUG: Currently it ignores JsonKey for binary secrets and returns the whole blob
		assert.Equal(t, "bin-value", resolved, "Should extract key from binary secret containing JSON")
	})

	t.Run("AwsSecretManager Binary Raw Preserves Whitespace", func(t *testing.T) {
		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-binary-raw")
		// No JsonKey

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)

		// Verify whitespace is preserved
		assert.Equal(t, "raw-value\n", resolved, "Should preserve trailing newline in binary secret")
	})
}
