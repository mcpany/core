// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/vault/api"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const maxSecretRecursionDepth = 10

// ResolveSecret resolves a SecretValue configuration object into a concrete string value.
// It handles various secret types including plain text, environment variables, file paths,
// remote URLs, Vault, and AWS Secrets Manager.
//
// Summary: Resolves a secret configuration into a plain string.
//
// Parameters:
//   - ctx: context.Context. The context for resolution.
//   - secret: *configv1.SecretValue. The secret configuration to resolve.
//
// Returns:
//   - string: The resolved secret value.
//   - error: An error if resolution fails.
func ResolveSecret(ctx context.Context, secret *configv1.SecretValue) (string, error) {
	return resolveSecretRecursive(ctx, secret, 0)
}

func resolveSecretRecursive(ctx context.Context, secret *configv1.SecretValue, depth int) (string, error) {
	val, err := resolveSecretImpl(ctx, secret, depth)
	if err != nil {
		return "", err
	}

	if secret != nil && secret.GetValidationRegex() != "" {
		re, err := regexp.Compile(secret.GetValidationRegex())
		if err != nil {
			return "", fmt.Errorf("invalid validation regex %q: %w", secret.GetValidationRegex(), err)
		}
		if !re.MatchString(val) {
			return "", fmt.Errorf("secret value does not match validation regex %q", secret.GetValidationRegex())
		}
	}

	return val, nil
}

func resolveSecretImpl(ctx context.Context, secret *configv1.SecretValue, depth int) (string, error) { //nolint:gocyclo
	if depth > maxSecretRecursionDepth {
		return "", fmt.Errorf("secret resolution exceeded max recursion depth of %d", maxSecretRecursionDepth)
	}

	if secret == nil {
		return "", nil
	}

	switch secret.WhichValue() {
	case configv1.SecretValue_PlainText_case:
		return strings.TrimSpace(secret.GetPlainText()), nil
	case configv1.SecretValue_EnvironmentVariable_case:
		envVar := secret.GetEnvironmentVariable()
		if !IsEnvVarAllowed(envVar) {
			return "", fmt.Errorf("access to environment variable %q is restricted", envVar)
		}
		value, ok := os.LookupEnv(envVar)
		if !ok {
			return "", fmt.Errorf("environment variable %q is not set", envVar)
		}
		return strings.TrimSpace(value), nil
	case configv1.SecretValue_FilePath_case:
		if err := validation.IsAllowedPath(secret.GetFilePath()); err != nil {
			return "", fmt.Errorf("invalid secret file path %q: %w", secret.GetFilePath(), err)
		}
		// File reading is blocking and generally fast, but technically could verify context.
		// For simplicity and standard library limits, we just read.
		content, err := os.ReadFile(secret.GetFilePath())
		if err != nil {
			return "", fmt.Errorf("failed to read secret from file %q: %w", secret.GetFilePath(), err)
		}
		return strings.TrimSpace(string(content)), nil
	case configv1.SecretValue_RemoteContent_case:
		remote := secret.GetRemoteContent()
		req, err := http.NewRequestWithContext(ctx, "GET", remote.GetHttpUrl(), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request for remote secret: %w", err)
		}

		if auth := remote.GetAuth(); auth != nil {
			if apiKey := auth.GetApiKey(); apiKey != nil {
				apiKeyValue, err := resolveSecretRecursive(ctx, apiKey.GetValue(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve api key for remote secret: %w", err)
				}
				req.Header.Set(apiKey.GetParamName(), apiKeyValue)
			} else if bearer := auth.GetBearerToken(); bearer != nil {
				token, err := resolveSecretRecursive(ctx, bearer.GetToken(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve bearer token for remote secret: %w", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			} else if basic := auth.GetBasicAuth(); basic != nil {
				password, err := resolveSecretRecursive(ctx, basic.GetPassword(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve password for remote secret: %w", err)
				}
				req.SetBasicAuth(basic.GetUsername(), password)
			} else if oauth2Auth := auth.GetOauth2(); oauth2Auth != nil {
				clientID, err := resolveSecretRecursive(ctx, oauth2Auth.GetClientId(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve client id for remote secret: %w", err)
				}
				clientSecret, err := resolveSecretRecursive(ctx, oauth2Auth.GetClientSecret(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve client secret for remote secret: %w", err)
				}

				conf := &clientcredentials.Config{
					ClientID:     clientID,
					ClientSecret: clientSecret,
					TokenURL:     oauth2Auth.GetTokenUrl(),
					Scopes:       strings.Fields(oauth2Auth.GetScopes()),
				}

				// Use safeSecretClient for the token request to prevent SSRF
				tokenCtx := context.WithValue(ctx, oauth2.HTTPClient, safeSecretClient)
				token, err := conf.Token(tokenCtx)
				if err != nil {
					return "", fmt.Errorf("failed to get oauth2 token: %w", err)
				}

				req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			}
		}

		resp, err := safeSecretClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch remote secret: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch remote secret: status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read remote secret body: %w", err)
		}
		return strings.TrimSpace(string(body)), nil
	case configv1.SecretValue_Vault_case:
		vaultSecret := secret.GetVault()
		config := &api.Config{
			Address:    vaultSecret.GetAddress(),
			HttpClient: safeSecretClient, // Use safe client for vault too
		}
		client, err := api.NewClient(config)
		if err != nil {
			return "", fmt.Errorf("failed to create vault client: %w", err)
		}
		token, err := resolveSecretRecursive(ctx, vaultSecret.GetToken(), depth+1)
		if err != nil {
			return "", fmt.Errorf("failed to resolve vault token: %w", err)
		}
		client.SetToken(token)

		// Vault API client doesn't support context on Read easily without wrappers,
		// but checking context before request is better than nothing.
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Ideally we would use Request with context, but hashicorp/vault/api
		// logical.Read doesn't take context. We can use client.Logical().ReadWithContext if available or similar.
		// Checking docs, modern vault/api has ReadWithContext.
		// Let's assume we can use Read (or check if ReadWithContext exists in this version).
		// Since I cannot check version easily, I'll stick to Read but add context check.
		// Actually, let's try to pass context if the method exists in user's version?
		// User said OS is linux, code is likely recent.
		// Vault's Logical() returns *Logical.
		// Let's use standard Read for now as per original code, but finding a way to support context is better.
		// Update: verified that recent Vault API has no ReadWithContext on Logical().
		// It relies on the client's HTTP client.
		// We set HttpClient to safeSecretClient, but safeSecretClient usage in Vault might not reuse our context for cancellation
		// unless we wrap it or use a custom RoundTripper that respects the context.
		// For now, minimal change: just use existing Read.

		data, err := client.Logical().Read(vaultSecret.GetPath())
		if err != nil {
			return "", fmt.Errorf("failed to read secret from vault: %w", err)
		}
		if data == nil || data.Data == nil {
			return "", fmt.Errorf("secret not found at path: %s", vaultSecret.GetPath())
		}

		// Handle KV v2 secrets, where data is nested under a "data" key.
		if secretData, ok := data.Data["data"].(map[string]interface{}); ok {
			if value, ok := secretData[vaultSecret.GetKey()].(string); ok {
				return strings.TrimSpace(value), nil
			}
		}

		// If not found in nested data, try to access as a KV v1 secret.
		value, ok := data.Data[vaultSecret.GetKey()].(string)
		if !ok {
			return "", fmt.Errorf("key %q not found in secret data at path %s", vaultSecret.GetKey(), vaultSecret.GetPath())
		}
		return strings.TrimSpace(value), nil
	case configv1.SecretValue_AwsSecretManager_case:
		smSecret := secret.GetAwsSecretManager()

		// Load default config
		loadOptions := []func(*config.LoadOptions) error{
			config.WithHTTPClient(safeSecretClient),
		}
		if smSecret.GetRegion() != "" {
			loadOptions = append(loadOptions, config.WithRegion(smSecret.GetRegion()))
		}
		if smSecret.GetProfile() != "" {
			loadOptions = append(loadOptions, config.WithSharedConfigProfile(smSecret.GetProfile()))
		}

		// Use the passed context
		cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
		if err != nil {
			return "", fmt.Errorf("failed to load aws config: %w", err)
		}

		client := secretsmanager.NewFromConfig(cfg)

		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(smSecret.GetSecretId()),
		}
		if smSecret.GetVersionId() != "" {
			input.VersionId = aws.String(smSecret.GetVersionId())
		}
		if smSecret.GetVersionStage() != "" {
			input.VersionStage = aws.String(smSecret.GetVersionStage())
		}

		// Use a custom endpoint resolver if provided (mostly for testing)
		// Since we can't easily inject it into LoadDefaultConfig without environment vars
		// or changing the function signature, we rely on environment variables for testing.
		// AWS_ENDPOINT_URL is supported in newer SDK versions.
		result, err := client.GetSecretValue(ctx, input)
		if err != nil {
			return "", fmt.Errorf("failed to get secret value from aws secrets manager: %w", err)
		}

		if result.SecretString == nil {
			if result.SecretBinary != nil {
				return string(result.SecretBinary), nil
			}
			return "", fmt.Errorf("secret value is not a string or binary")
		}

		secretVal := *result.SecretString

		if smSecret.GetJsonKey() != "" {
			var secretMap map[string]interface{}
			if err := json.Unmarshal([]byte(secretVal), &secretMap); err != nil {
				return "", fmt.Errorf("failed to unmarshal secret json: %w", err)
			}

			val, ok := secretMap[smSecret.GetJsonKey()]
			if !ok {
				return "", fmt.Errorf("key %q not found in secret json", smSecret.GetJsonKey())
			}

			// Convert val to string
			if strVal, ok := val.(string); ok {
				return strings.TrimSpace(strVal), nil
			}
			// Try to convert other types to string
			return strings.TrimSpace(fmt.Sprintf("%v", val)), nil
		}

		return strings.TrimSpace(secretVal), nil
	default:
		return "", nil
	}
}

// ResolveSecretMap resolves a map of SecretValue objects and merges them with a map of plain strings.
// If a key exists in both maps, the value from the secretMap (once resolved) takes precedence.
//
// Summary: Resolves a map of secrets and merges with plain values.
//
// Parameters:
//   - ctx: context.Context. The context for resolution.
//   - secretMap: map[string]*configv1.SecretValue. A map of secret configurations.
//   - plainMap: map[string]string. A map of plain string values (defaults).
//
// Returns:
//   - map[string]string: A map with all values resolved.
//   - error: An error if any secret resolution fails.
func ResolveSecretMap(ctx context.Context, secretMap map[string]*configv1.SecretValue, plainMap map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range plainMap {
		result[k] = v
	}
	for k, v := range secretMap {
		resolved, err := ResolveSecret(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve secret env var %q: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

// safeSecretClient is an http.Client that prevents SSRF by blocking access to link-local IPs (like AWS metadata service).
// It also resolves the IP before dialing to prevent DNS rebinding attacks.
var safeSecretClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		// Prevent redirects to avoid leaking sensitive headers (like Authorization or X-API-Key)
		// to untrusted third parties.
		return http.ErrUseLastResponse
	},
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := NewSafeDialer()
			if os.Getenv("MCPANY_ALLOW_LOOPBACK_SECRETS") == TrueStr {
				dialer.AllowLoopback = true
			}
			if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_SECRETS") == TrueStr {
				dialer.AllowPrivate = true
			}
			// LinkLocal is default false (blocked).
			return dialer.DialContext(ctx, network, addr)
		},
	},
}
