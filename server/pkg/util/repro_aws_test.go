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
			w.WriteHeader(http.StatusOK)

            // Secret content is JSON, but stored as binary
            secretContent := []byte(`{"bin-key": "bin-value"}`)

            // AWS requires binary fields to be base64 encoded in JSON response
            // SDK v2 might expect raw bytes if using specific protocol?
            // AWS Secrets Manager uses AWS JSON 1.1 protocol.
            // In JSON, Blob types are Base64 encoded strings.

			// Construct response.
            // We set SecretBinary, not SecretString.
			resp := map[string]interface{}{
				"Name":         "my-binary-secret",
				"SecretBinary": secretContent,
			}
            // Wait, encoding []byte to JSON automatically does Base64 encoding in Go.
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
}
