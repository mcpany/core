// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret(t *testing.T) {
	t.Run("nil secret", func(t *testing.T) {
		resolved, err := ResolveSecret(nil)
		assert.NoError(t, err)
		assert.Empty(t, resolved)
	})

	t.Run("unknown secret type", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Empty(t, resolved)
	})

	t.Run("PlainText", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetPlainText("my-secret")
		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-secret", resolved)
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		t.Setenv("MY_SECRET_ENV", "my-env-secret")
		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("MY_SECRET_ENV")
		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-env-secret", resolved)
	})

	t.Run("EnvironmentVariable not set", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("MY_SECRET_ENV_NOT_SET")
		_, err := ResolveSecret(secret)
		assert.Error(t, err)
	})

	t.Run("FilePath", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "secret")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(tmpfile.Name()) }()

		_, err = tmpfile.WriteString("my-file-secret")
		assert.NoError(t, err)
		_ = tmpfile.Close()

		secret := &configv1.SecretValue{}
		secret.SetFilePath(tmpfile.Name())
		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-file-secret", resolved)
	})

	t.Run("FilePath not found", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetFilePath("non-existent-file")
		_, err := ResolveSecret(secret)
		assert.Error(t, err)
	})

	t.Run("RemoteContent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, "my-remote-secret")
		}))
		defer server.Close()

		secret := &configv1.SecretValue{}
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		secret.SetRemoteContent(remoteContent)

		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent with API Key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "my-api-key", r.Header.Get("X-API-Key"))
			_, _ = fmt.Fprint(w, "my-remote-secret")
		}))
		defer server.Close()

		apiKeySecret := &configv1.SecretValue{}
		apiKeySecret.SetPlainText("my-api-key")

		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("X-API-Key")
		apiKeyAuth.SetApiKey(apiKeySecret)

		auth := &configv1.Authentication{}
		auth.SetApiKey(apiKeyAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent with Bearer Token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer my-bearer-token", r.Header.Get("Authorization"))
			_, _ = fmt.Fprint(w, "my-remote-secret")
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-bearer-token")

		bearerTokenAuth := &configv1.UpstreamBearerTokenAuth{}
		bearerTokenAuth.SetToken(tokenSecret)

		auth := &configv1.Authentication{}
		auth.SetBearerToken(bearerTokenAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent with Basic Auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "my-user", username)
			assert.Equal(t, "my-password", password)
			_, _ = fmt.Fprint(w, "my-remote-secret")
		}))
		defer server.Close()

		passwordSecret := &configv1.SecretValue{}
		passwordSecret.SetPlainText("my-password")

		basicAuth := &configv1.UpstreamBasicAuth{}
		basicAuth.SetUsername("my-user")
		basicAuth.SetPassword(passwordSecret)

		auth := &configv1.Authentication{}
		auth.SetBasicAuth(basicAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := ResolveSecret(secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent with bad request", func(t *testing.T) {
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("bad-url")
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(secret)
		assert.Error(t, err)
	})

	t.Run("RemoteContent with status not ok", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(secret)
		assert.Error(t, err)
	})

	t.Run("RemoteContent with read error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))
		defer server.Close()

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(secret)
		assert.Error(t, err)
	})
}
