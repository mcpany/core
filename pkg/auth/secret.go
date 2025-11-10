/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ResolveSecretValue retrieves the string value from a SecretValue message,
// handling different sources like plain text, environment variables, and files.
func ResolveSecretValue(secret *configv1.SecretValue) (string, error) {
	if secret == nil {
		return "", nil
	}
	switch secret.WhichValue() {
	case configv1.SecretValue_PlainText_case:
		return secret.GetPlainText(), nil
	case configv1.SecretValue_EnvironmentVariable_case:
		val := os.Getenv(secret.GetEnvironmentVariable())
		if val == "" {
			return "", fmt.Errorf("environment variable %q is not set", secret.GetEnvironmentVariable())
		}
		return val, nil
	case configv1.SecretValue_FilePath_case:
		content, err := os.ReadFile(secret.GetFilePath())
		if err != nil {
			return "", fmt.Errorf("could not read secret from file %q: %w", secret.GetFilePath(), err)
		}
		return strings.TrimSpace(string(content)), nil
	case configv1.SecretValue_RemoteContent_case:
		return "", errors.New("remote_content secret value is not implemented yet")
	default:
		return "", nil
	}
}
