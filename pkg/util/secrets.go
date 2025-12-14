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

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/vault/api"
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

		resp, err := http.DefaultClient.Do(req)
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
		secretData, ok := data.Data["data"].(map[string]interface{})
		if !ok {
			// If not nested, try to access as a KV v1 secret.
			value, ok := data.Data[vaultSecret.GetKey()].(string)
			if !ok {
				return "", fmt.Errorf("key %q not found in secret at path %s, and secret format is not standard KV v1 or v2", vaultSecret.GetKey(), vaultSecret.GetPath())
			}
			return value, nil
		}

		// It's a KV v2 secret
		value, ok := secretData[vaultSecret.GetKey()].(string)
		if !ok {
			return "", fmt.Errorf("key %q not found in secret data at path %s", vaultSecret.GetKey(), vaultSecret.GetPath())
		}
		return value, nil
	default:
		return "", nil
	}
}
