// Copyright 2025 Author(s) of MCP Any
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

func TestResolveSecret_AwsSecretManager(t *testing.T) {
	util.ClearSecretCache()
	// Mock AWS Secrets Manager Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assert.Equal(t, "POST", r.Method)
		// assert.Equal(t, "application/x-amz-json-1.1", r.Header.Get("Content-Type"))

		// Simple routing based on target
		target := r.Header.Get("X-Amz-Target")
		if target == "secretsmanager.GetSecretValue" {
			w.WriteHeader(http.StatusOK)

			// Construct response.
			// Just return the standard body regardless of SecretId, this is a simple mock
			resp := map[string]string{
				"Name":         "my-secret",
				"SecretString": `{"my-key": "my-aws-value", "other-key": 123}`,
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
	// This is the magic env var for AWS SDK v2 to override endpoint
	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	t.Run("AwsSecretManager full JSON", func(t *testing.T) {
		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-secret")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)
		util.ClearSecretCache()

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Contains(t, resolved, "my-aws-value")
		assert.Contains(t, resolved, "other-key")
	})

	t.Run("AwsSecretManager with key", func(t *testing.T) {
		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-secret")
		awsSecret.SetJsonKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)

		util.ClearSecretCache()

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-aws-value", resolved)

	})

	t.Run("AwsSecretManager caching", func(t *testing.T) {
		util.ClearSecretCache()

		callCount := 0
		cacheServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
			resp := map[string]string{
				"Name":         "my-cached-secret",
				"SecretString": `{"cached-key": "cached-value"}`,
			}

			_ = json.NewEncoder(w).Encode(resp)
			return
		}))
		defer cacheServer.Close()
		t.Setenv("AWS_ENDPOINT_URL", cacheServer.URL)

		awsCachedSecret := &configv1.AwsSecretManagerSecret{}
		awsCachedSecret.SetSecretId("my-cached-secret")
		awsCachedSecret.SetJsonKey("cached-key")
		cachedSecret := &configv1.SecretValue{}
		cachedSecret.SetAwsSecretManager(awsCachedSecret)

		res1, err1 := util.ResolveSecret(context.Background(), cachedSecret)
		assert.NoError(t, err1)
		assert.Equal(t, "cached-value", res1)

		res2, err2 := util.ResolveSecret(context.Background(), cachedSecret)
		assert.NoError(t, err2)
		assert.Equal(t, "cached-value", res2)

		assert.Equal(t, 1, callCount, "Should have only called AWS once due to cache")
	})

	t.Run("AwsSecretManager with non-string key", func(t *testing.T) {
		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-secret")
		awsSecret.SetJsonKey("other-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)

		util.ClearSecretCache()

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "123", resolved)
	})

	t.Run("AwsSecretManager failure", func(t *testing.T) {
		// Create a separate server for failure case to avoid messing up the shared one
		failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer failServer.Close()

		// Override endpoint for this test only
		t.Setenv("AWS_ENDPOINT_URL", failServer.URL)

		awsSecret := &configv1.AwsSecretManagerSecret{}
		awsSecret.SetSecretId("my-fail-secret")
		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(awsSecret)
		util.ClearSecretCache()

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})
}
