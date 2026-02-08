package util_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
	// t.Setenv("MCPANY_FILE_PATH_ALLOW_LIST", os.TempDir())
	validation.SetAllowedPaths([]string{os.TempDir()})
	t.Cleanup(func() { validation.SetAllowedPaths(nil) })

	t.Run("nil secret", func(t *testing.T) {
		resolved, err := util.ResolveSecret(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, "", resolved)
	})

	t.Run("unknown secret type", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Empty(t, resolved)
	})

	t.Run("PlainText", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetPlainText("my-secret")
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-secret", resolved)
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		t.Setenv("MY_SECRET_ENV", "my-env-secret")
		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("MY_SECRET_ENV")
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-env-secret", resolved)
	})

	t.Run("EnvironmentVariable not set", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("MY_SECRET_ENV_NOT_SET")
		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable \"MY_SECRET_ENV_NOT_SET\" is not set")
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
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-file-secret", resolved)
	})

	t.Run("FilePath not found", func(t *testing.T) {
		secret := &configv1.SecretValue{}
		secret.SetFilePath("non-existent-file")
		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})

	t.Run("FilePath with newline", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "secret_with_newline")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(tmpfile.Name()) }()

		_, err = tmpfile.WriteString("my-file-secret-newline\n")
		assert.NoError(t, err)
		_ = tmpfile.Close()

		secret := &configv1.SecretValue{}
		secret.SetFilePath(tmpfile.Name())
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-file-secret-newline", resolved)
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

		resolved, err := util.ResolveSecret(context.Background(), secret)
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

		apiKeyAuth := &configv1.APIKeyAuth{}
		apiKeyAuth.SetParamName("X-API-Key")
		apiKeyAuth.SetValue(apiKeySecret)

		auth := &configv1.Authentication{}
		auth.SetApiKey(apiKeyAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := util.ResolveSecret(context.Background(), secret)
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

		bearerTokenAuth := &configv1.BearerTokenAuth{}
		bearerTokenAuth.SetToken(tokenSecret)

		auth := &configv1.Authentication{}
		auth.SetBearerToken(bearerTokenAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := util.ResolveSecret(context.Background(), secret)
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

		basicAuth := &configv1.BasicAuth{}
		basicAuth.SetUsername("my-user")
		basicAuth.SetPassword(passwordSecret)

		auth := &configv1.Authentication{}
		auth.SetBasicAuth(basicAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(server.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent with OAuth2", func(t *testing.T) {
		// Mock Token Server
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			// Check credentials
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "my-client-id", username)
			assert.Equal(t, "my-client-secret", password)

			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token": "my-access-token", "token_type": "Bearer", "expires_in": 3600}`)
		}))
		defer tokenServer.Close()

		// Mock Resource Server
		resourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer my-access-token", r.Header.Get("Authorization"))
			_, _ = fmt.Fprint(w, "my-remote-secret-oauth")
		}))
		defer resourceServer.Close()

		clientIDSecret := &configv1.SecretValue{}
		clientIDSecret.SetPlainText("my-client-id")

		clientSecretSecret := &configv1.SecretValue{}
		clientSecretSecret.SetPlainText("my-client-secret")

		oauth2Auth := &configv1.OAuth2Auth{}
		oauth2Auth.SetClientId(clientIDSecret)
		oauth2Auth.SetClientSecret(clientSecretSecret)
		oauth2Auth.SetTokenUrl(tokenServer.URL)
		oauth2Auth.SetScopes("read")

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(resourceServer.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret-oauth", resolved)
	})

	t.Run("RemoteContent with bad request", func(t *testing.T) {
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("bad-url")
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
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

		_, err := util.ResolveSecret(context.Background(), secret)
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

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})
}

func TestResolveSecret_Vault(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	t.Run("Vault", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/secret/data/my-app/db", r.URL.Path)
			assert.Equal(t, "my-vault-token", r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"data": {"data": {"my-key": "my-vault-secret"}}}`)
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-vault-token")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress(server.URL)
		vaultSecret.SetToken(tokenSecret)
		vaultSecret.SetPath("secret/data/my-app/db")
		vaultSecret.SetKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-vault-secret", resolved)
	})

	t.Run("Vault secret not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-vault-token")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress(server.URL)
		vaultSecret.SetToken(tokenSecret)
		vaultSecret.SetPath("secret/data/my-app/db")
		vaultSecret.SetKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})

	t.Run("Vault key not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"data": {"data": {"another-key": "another-value"}}}`)
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-vault-token")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress(server.URL)
		vaultSecret.SetToken(tokenSecret)
		vaultSecret.SetPath("secret/data/my-app/db")
		vaultSecret.SetKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})

	t.Run("Vault KV v1", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/secret/my-app/db", r.URL.Path)
			assert.Equal(t, "my-vault-token", r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"data": {"my-key": "my-vault-secret-v1"}}`)
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-vault-token")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress(server.URL)
		vaultSecret.SetToken(tokenSecret)
		vaultSecret.SetPath("secret/my-app/db")
		vaultSecret.SetKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-vault-secret-v1", resolved)
	})

	t.Run("Vault KV v1 with data key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/secret/my-app/db", r.URL.Path)
			assert.Equal(t, "my-vault-token", r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusOK)
			// Simulating a KV v1 secret that has a key named "data" which is a map (object),
			// alongside the actual key we want ("my-key").
			_, _ = fmt.Fprint(w, `{"data": {"my-key": "my-vault-secret-v1", "data": {"some": "nested-value"}}}`)
		}))
		defer server.Close()

		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetPlainText("my-vault-token")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress(server.URL)
		vaultSecret.SetToken(tokenSecret)
		vaultSecret.SetPath("secret/my-app/db")
		vaultSecret.SetKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-vault-secret-v1", resolved)
	})

	t.Run("Secret resolution recursion limit", func(t *testing.T) {
		// Create a circular dependency:
		// secretA's token is secretB
		// secretB's token is secretA
		secretA := &configv1.SecretValue{}
		vaultA := &configv1.VaultSecret{}
		vaultA.SetAddress("http://fake-vault")
		vaultA.SetPath("secret/a")
		vaultA.SetKey("key")
		secretA.SetVault(vaultA)

		secretB := &configv1.SecretValue{}
		vaultB := &configv1.VaultSecret{}
		vaultB.SetAddress("http://fake-vault")
		vaultB.SetPath("secret/b")
		vaultB.SetKey("key")
		secretB.SetVault(vaultB)

		// Create the cycle
		vaultA.SetToken(secretB)
		vaultB.SetToken(secretA)

		_, err := util.ResolveSecret(context.Background(), secretA)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret resolution exceeded max recursion depth")
	})

	t.Run("Context Cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		secret := &configv1.SecretValue{}
		// Use Vault type to trigger the context check
		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String("http://127.0.0.1:8200"),
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
			Path: proto.String("secret/foo"),
			Key:  proto.String("bar"),
		}.Build()
		vaultSecret.SetPath("secret/foo")
		vaultSecret.SetKey("bar")
		secret.SetVault(vaultSecret)

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("AwsSecretManager failure", func(t *testing.T) {
		// Expect error because no credentials/network
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
	})
}

func TestResolveSecretMap(t *testing.T) {
	t.Run("Merge", func(t *testing.T) {
		secretMap := map[string]*configv1.SecretValue{
			"SECRET_VAR": configv1.SecretValue_builder{
				PlainText: proto.String("secret_value"),
			}.Build(),
		}
		plainMap := map[string]string{
			"PLAIN_VAR":  "plain_value",
			"SECRET_VAR": "will_be_overridden",
		}

		resolved, err := util.ResolveSecretMap(context.Background(), secretMap, plainMap)
		assert.NoError(t, err)
		assert.Equal(t, "plain_value", resolved["PLAIN_VAR"])
		assert.Equal(t, "secret_value", resolved["SECRET_VAR"])
	})

	t.Run("ResolveError", func(t *testing.T) {
		secretMap := map[string]*configv1.SecretValue{
			"SECRET_VAR": configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("NON_EXISTENT_VAR"),
			}.Build(),
		}
		plainMap := map[string]string{}

		_, err := util.ResolveSecretMap(context.Background(), secretMap, plainMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR\" is not set")
	})

	t.Run("resolve secret map", func(t *testing.T) {
		secretMap := map[string]*configv1.SecretValue{
			"KEY1": configv1.SecretValue_builder{
				PlainText: proto.String("test-value"),
			}.Build(),
			"KEY3": configv1.SecretValue_builder{
				PlainText: proto.String("test-secret-override"),
			}.Build(),
		}
		plainMap := map[string]string{
			"KEY2": "test-plain",
			"KEY3": "test-plain-override",
		}
		resolved, err := util.ResolveSecretMap(context.Background(), secretMap, plainMap)
		assert.NoError(t, err)
		assert.Equal(t, "test-value", resolved["KEY1"])
		assert.Equal(t, "test-plain", resolved["KEY2"])
		assert.Equal(t, "test-secret-override", resolved["KEY3"])
	})

	t.Run("resolve secret map error", func(t *testing.T) {
		secretMap := map[string]*configv1.SecretValue{}
		plainMap := map[string]string{}
		secretMap["KEY4"] = configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR"),
		}.Build()
		_, err := util.ResolveSecretMap(context.Background(), secretMap, plainMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR\" is not set")
	})
}
