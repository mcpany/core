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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/vault/api"
	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

const maxSecretRecursionDepth = 10

// ResolveSecret resolves a SecretValue into a string.
func ResolveSecret(secret *configv1.SecretValue) (string, error) {
	return resolveSecretRecursive(secret, 0)
}

func resolveSecretRecursive(secret *configv1.SecretValue, depth int) (string, error) {
	if depth > maxSecretRecursionDepth {
		return "", fmt.Errorf("secret resolution exceeded max recursion depth of %d", maxSecretRecursionDepth)
	}

	if secret == nil {
		return "", nil
	}

	switch secret.WhichValue() {
	case configv1.SecretValue_PlainText_case:
		return secret.GetPlainText(), nil
	case configv1.SecretValue_EnvironmentVariable_case:
		envVar := secret.GetEnvironmentVariable()
		value := os.Getenv(envVar)
		if value == "" {
			return "", fmt.Errorf("environment variable %q is not set", envVar)
		}
		return value, nil
	case configv1.SecretValue_FilePath_case:
		if err := validation.IsSecurePath(secret.GetFilePath()); err != nil {
			return "", fmt.Errorf("invalid secret file path %q: %w", secret.GetFilePath(), err)
		}
		content, err := os.ReadFile(secret.GetFilePath())
		if err != nil {
			return "", fmt.Errorf("failed to read secret from file %q: %w", secret.GetFilePath(), err)
		}
		return string(content), nil
	case configv1.SecretValue_RemoteContent_case:
		remote := secret.GetRemoteContent()
		req, err := http.NewRequest("GET", remote.GetHttpUrl(), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request for remote secret: %w", err)
		}

		if auth := remote.GetAuth(); auth != nil {
			if apiKey := auth.GetApiKey(); apiKey != nil {
				apiKeyValue, err := resolveSecretRecursive(apiKey.GetApiKey(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve api key for remote secret: %w", err)
				}
				req.Header.Set(apiKey.GetHeaderName(), apiKeyValue)
			} else if bearer := auth.GetBearerToken(); bearer != nil {
				token, err := resolveSecretRecursive(bearer.GetToken(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve bearer token for remote secret: %w", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			} else if basic := auth.GetBasicAuth(); basic != nil {
				password, err := resolveSecretRecursive(basic.GetPassword(), depth+1)
				if err != nil {
					return "", fmt.Errorf("failed to resolve password for remote secret: %w", err)
				}
				req.SetBasicAuth(basic.GetUsername(), password)
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
		return string(body), nil
	case configv1.SecretValue_Vault_case:
		vaultSecret := secret.GetVault()
		config := &api.Config{
			Address: vaultSecret.GetAddress(),
		}
		client, err := api.NewClient(config)
		if err != nil {
			return "", fmt.Errorf("failed to create vault client: %w", err)
		}
		token, err := resolveSecretRecursive(vaultSecret.GetToken(), depth+1)
		if err != nil {
			return "", fmt.Errorf("failed to resolve vault token: %w", err)
		}
		client.SetToken(token)
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
				return value, nil
			}
		}

		// If not found in nested data, try to access as a KV v1 secret.
		value, ok := data.Data[vaultSecret.GetKey()].(string)
		if !ok {
			return "", fmt.Errorf("key %q not found in secret data at path %s", vaultSecret.GetKey(), vaultSecret.GetPath())
		}
		return value, nil
	case configv1.SecretValue_AwsSecretManager_case:
		smSecret := secret.GetAwsSecretManager()

		// Load default config
		loadOptions := []func(*config.LoadOptions) error{}
		if smSecret.GetRegion() != "" {
			loadOptions = append(loadOptions, config.WithRegion(smSecret.GetRegion()))
		}
		if smSecret.GetProfile() != "" {
			loadOptions = append(loadOptions, config.WithSharedConfigProfile(smSecret.GetProfile()))
		}

		cfg, err := config.LoadDefaultConfig(context.TODO(), loadOptions...)
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

		result, err := client.GetSecretValue(context.TODO(), input)
		if err != nil {
			return "", fmt.Errorf("failed to get secret value from aws secrets manager: %w", err)
		}

		if result.SecretString == nil {
			// Handle binary secret? For now, we only support string.
			return "", fmt.Errorf("secret value is not a string (binary secrets not supported)")
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
				return strVal, nil
			}
			// Try to convert other types to string
			return fmt.Sprintf("%v", val), nil
		}

		return secretVal, nil
	default:
		return "", nil
	}
}

// ResolveSecretMap resolves a map of secrets and merges them with a plain map.
// Secrets in the secretMap take precedence over values in the plainMap.
func ResolveSecretMap(secretMap map[string]*configv1.SecretValue, plainMap map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range plainMap {
		result[k] = v
	}
	for k, v := range secretMap {
		resolved, err := ResolveSecret(v)
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
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
					return nil, fmt.Errorf("blocked link-local IP: %s", ip)
				}
				if os.Getenv("MCPANY_ALLOW_LOOPBACK_SECRETS") != "true" && ip.IsLoopback() {
					return nil, fmt.Errorf("blocked loopback IP: %s", ip)
				}
				if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_SECRETS") != "true" && ip.IsPrivate() {
					return nil, fmt.Errorf("blocked private IP: %s", ip)
				}
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no IPs resolved for %s", host)
			}
			// Use the first resolved IP to avoid DNS rebinding race conditions
			var d net.Dialer
			return d.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
	},
}
