package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go-v2/aws"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsSecretManager_JSONKey_Coverage(t *testing.T) {
	// Mock an AWS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Target") == "secretsmanager.GetSecretValue" {
			w.WriteHeader(http.StatusOK)
			// Return a JSON structure that secretsmanager returns
			// It has SecretString inside.
			w.Write([]byte(`{"SecretString": "{\"my-key\":\"my-value\", \"other-key\": 123}"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// Custom resolver to hit our local server instead of actual AWS metadata
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               server.URL,
			SigningRegion:     "us-east-1",
		}, nil
	})

	// Create context with the custom resolver for load config
	ctx := context.WithValue(context.Background(), "aws-endpoint-resolver", customResolver)

	t.Run("Valid JSON Key String", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("my-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(ctx, secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-value", resolved)
	})

	t.Run("Valid JSON Key Int", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("other-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(ctx, secret)
		assert.NoError(t, err)
		assert.Equal(t, "123", resolved)
	})

	t.Run("Missing JSON Key", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")
		smSecret.SetJsonKey("missing-key")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		_, err := ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in secret json")
	})
}

func TestAwsSecretManager_BadJSON_Coverage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"SecretString": "not-a-json"}`))
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               server.URL,
			SigningRegion:     "us-east-1",
		}, nil
	})
	ctx := context.WithValue(context.Background(), "aws-endpoint-resolver", customResolver)

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-secret")
	smSecret.SetRegion("us-east-1")
	smSecret.SetJsonKey("my-key")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	_, err := ResolveSecret(ctx, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal secret json")
}

func TestAwsSecretManager_Binary_Coverage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Base64 encoded "binary-data"
		w.Write([]byte(`{"SecretBinary": "YmluYXJ5LWRhdGE="}`))
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               server.URL,
			SigningRegion:     "us-east-1",
		}, nil
	})
	ctx := context.WithValue(context.Background(), "aws-endpoint-resolver", customResolver)

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-secret")
	smSecret.SetRegion("us-east-1")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	resolved, err := ResolveSecret(ctx, secret)
	assert.NoError(t, err)
	assert.Equal(t, "binary-data", resolved)
}

func TestAwsSecretManager_Empty_Coverage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               server.URL,
			SigningRegion:     "us-east-1",
		}, nil
	})
	ctx := context.WithValue(context.Background(), "aws-endpoint-resolver", customResolver)

	smSecret := &configv1.AwsSecretManagerSecret{}
	smSecret.SetSecretId("my-secret")
	smSecret.SetRegion("us-east-1")

	secret := &configv1.SecretValue{}
	secret.SetAwsSecretManager(smSecret)

	_, err := ResolveSecret(ctx, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret value is not a string or binary")
}

func TestResolveSecret_RemoteContent_Coverage(t *testing.T) {
	t.Run("RemoteContent auth token failures", func(t *testing.T) {
		// Mock an HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

		tests := []struct{
			name string
			auth *configv1.Authentication
			errMsg string
		}{
			{
				name: "ApiKey failure",
				auth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						ParamName: proto.String("key"),
						Value: configv1.SecretValue_builder{
							EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_API_KEY"),
						}.Build(),
					}.Build(),
				}.Build(),
				errMsg: "failed to resolve api key for remote secret",
			},
			{
				name: "Bearer failure",
				auth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_BEARER"),
						}.Build(),
					}.Build(),
				}.Build(),
				errMsg: "failed to resolve bearer token for remote secret",
			},
			{
				name: "BasicAuth failure",
				auth: configv1.Authentication_builder{
					BasicAuth: configv1.BasicAuth_builder{
						Username: proto.String("user"),
						Password: configv1.SecretValue_builder{
							EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_PASSWORD"),
						}.Build(),
					}.Build(),
				}.Build(),
				errMsg: "failed to resolve password for remote secret",
			},
			{
				name: "OAuth2 client ID failure",
				auth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						ClientId: configv1.SecretValue_builder{
							EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_CLIENT_ID"),
						}.Build(),
					}.Build(),
				}.Build(),
				errMsg: "failed to resolve client id for remote secret",
			},
			{
				name: "OAuth2 client secret failure",
				auth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						ClientId: configv1.SecretValue_builder{
							PlainText: proto.String("client-id"),
						}.Build(),
						ClientSecret: configv1.SecretValue_builder{
							EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_CLIENT_SECRET"),
						}.Build(),
					}.Build(),
				}.Build(),
				errMsg: "failed to resolve client secret for remote secret",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				remoteContent := configv1.RemoteContent_builder{
					HttpUrl: proto.String(server.URL),
					Auth: tc.auth,
				}.Build()
				secret := configv1.SecretValue_builder{
					RemoteContent: remoteContent,
				}.Build()

				_, err := ResolveSecret(context.Background(), secret)
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			})
		}
	})

	t.Run("RemoteContent Request creation failure", func(t *testing.T) {
		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("://bad-url"),
		}.Build()
		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create request for remote secret")
		}
	})
}

func TestVaultSecret_Coverage(t *testing.T) {
	t.Run("VaultSecret Token Resolve Failure", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String("http://127.0.0.1:8200"),
			Token: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("MISSING_ENV_VAR_FOR_VAULT_TOKEN"),
			}.Build(),
			Path: proto.String("secret/foo"),
			Key:  proto.String("bar"),
		}.Build()

		secret := configv1.SecretValue_builder{
			Vault: vaultSecret,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to resolve vault token")
		}
	})

	t.Run("VaultSecret Create Client Failure", func(t *testing.T) {
		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String("://bad-address"),
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
			Path: proto.String("secret/foo"),
			Key:  proto.String("bar"),
		}.Build()

		secret := configv1.SecretValue_builder{
			Vault: vaultSecret,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create vault client")
		}
	})
}

func TestResolveSecret_RegexCache_Coverage(t *testing.T) {
	t.Run("Invalid Regex", func(t *testing.T) {
		pt := configv1.SecretValue_builder{
			PlainText: proto.String("secret_pt"),
			ValidationRegex: proto.String("[unclosed"),
		}.Build()

		_, err := ResolveSecret(context.Background(), pt)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "invalid validation regex")
		}
	})

	t.Run("Disallowed Env Var", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_RESTRICTED_ENV_VARS", "false")
		t.Setenv("MCPANY_STRICT_ENV_MODE", "true")
		os.Setenv("RESTRICTED_VAR_AWS_ACCESS_KEY", "dummy")
		defer os.Unsetenv("RESTRICTED_VAR_AWS_ACCESS_KEY")
		secret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("RESTRICTED_VAR_AWS_ACCESS_KEY"),
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "access to environment variable")
		}
	})
}

func TestVaultSecret_DataNil_Coverage(t *testing.T) {
	t.Run("VaultSecret nil data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`)) // no data field
		}))
		defer server.Close()

		t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String(server.URL),
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
			Path: proto.String("secret/foo"),
			Key:  proto.String("bar"),
		}.Build()

		secret := configv1.SecretValue_builder{
			Vault: vaultSecret,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "secret not found at path")
		}
	})
}

func TestAwsSecretManager_BadConfig_Coverage(t *testing.T) {
	// Attempting to resolve with an invalid profile will cause LoadDefaultConfig to fail
	smSecret := configv1.AwsSecretManagerSecret_builder{
		SecretId: proto.String("my-secret"),
		Region: proto.String("us-east-1"),
		Profile: proto.String("invalid_profile_name"),
	}.Build()

	secret := configv1.SecretValue_builder{
		AwsSecretManager: smSecret,
	}.Build()

	_, err := ResolveSecret(context.Background(), secret)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to load aws config")
	}
}

func TestVaultSecret_ReadError_Coverage(t *testing.T) {
	t.Run("VaultSecret read error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// return error
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String(server.URL),
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
			Path: proto.String("secret/foo"),
			Key:  proto.String("bar"),
		}.Build()

		secret := configv1.SecretValue_builder{
			Vault: vaultSecret,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to read secret from vault")
		}
	})
}

func TestResolveSecret_FilePath_Coverage(t *testing.T) {
	t.Run("Invalid FilePath", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			FilePath: proto.String("../../../../etc/passwd"),
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "invalid secret file path")
		}
	})
}

func TestAwsSecretManager_Options2_Coverage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"SecretString": "my-value"}`))
	}))
	defer server.Close()

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               server.URL,
			SigningRegion:     "us-east-1",
		}, nil
	})
	ctx := context.WithValue(context.Background(), "aws-endpoint-resolver", customResolver)

	t.Run("With VersionId and VersionStage", func(t *testing.T) {
		smSecret := &configv1.AwsSecretManagerSecret{}
		smSecret.SetSecretId("my-secret")
		smSecret.SetRegion("us-east-1")
		smSecret.SetVersionId("")
		smSecret.SetVersionStage("")

		secret := &configv1.SecretValue{}
		secret.SetAwsSecretManager(smSecret)

		resolved, err := ResolveSecret(ctx, secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-value", resolved)
	})
}
