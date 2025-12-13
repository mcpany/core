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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/consts"
)

// SanitizeID sanitizes a slice of strings to create a single valid identifier.
//
// It processes each string in the input slice `ids` to make it conform to a set of rules,
// ensuring that the resulting identifier is safe and consistent for use in various contexts.
// The sanitization process involves:
//  1. Removing any characters that are not alphanumeric, underscore, or hyphen.
//  2. Truncating the string if it exceeds a specified maximum length.
//  3. Optionally, appending a hash of the original string to ensure uniqueness,
//     especially when truncation occurs or when illegal characters are present.
//
// After sanitizing each string individually, it joins them with a "." separator to form
// the final identifier.
//
// Parameters:
//
//	ids: A slice of strings to be sanitized and joined. Each element of the slice
//	     represents a part of the final identifier.
//	alwaysAppendHash: A boolean that, if true, forces a hash to be appended to each
//	                  sanitized string, regardless of whether it was modified.
//	maxSanitizedPrefixLength: The maximum allowed length for the sanitized prefix of each
//	                          string before a hash is appended.
//	hashLength: The desired length of the hexadecimal hash to be appended.
//
// Returns:
//
//	A single string representing the sanitized and joined identifier.
func SanitizeID(ids []string, alwaysAppendHash bool, maxSanitizedPrefixLength, hashLength int) (string, error) {
	var sanitizedIDs []string
	for _, id := range ids {
		if id == "" {
			return "", fmt.Errorf("id cannot be empty")
		}

		// Create the hash
		h := sha256.New()
		h.Write([]byte(id))
		hash := hex.EncodeToString(h.Sum(nil))
		if hashLength > 0 && hashLength < len(hash) {
			hash = hash[:hashLength]
		}

		// Sanitize and create the prefix
		sanitizedPrefix := nonWordChars.ReplaceAllString(id, "")

		// Check if we need to append the hash
		appendHash := alwaysAppendHash ||
			len(sanitizedPrefix) != len(id) ||
			len(sanitizedPrefix) > maxSanitizedPrefixLength

		if len(sanitizedPrefix) > maxSanitizedPrefixLength {
			sanitizedPrefix = sanitizedPrefix[:maxSanitizedPrefixLength]
		}

		if appendHash {
			if sanitizedPrefix == "" {
				sanitizedPrefix = "id"
			}
			sanitizedIDs = append(sanitizedIDs, fmt.Sprintf("%s_%s", sanitizedPrefix, hash))
		} else {
			sanitizedIDs = append(sanitizedIDs, sanitizedPrefix)
		}
	}

	return strings.Join(sanitizedIDs, "."), nil
}

// SanitizeServiceName sanitizes the given service name.
// It ensures that the name is a valid identifier by removing disallowed characters
// and appending a hash if the name is too long or contains illegal characters.
// This function calls SanitizeID with alwaysAppendHash set to false.
func SanitizeServiceName(name string) (string, error) {
	return SanitizeID([]string{name}, false, maxSanitizedPrefixLength, hashLength)
}

// SanitizeToolName sanitizes the given tool name.
// It ensures that the name is a valid identifier by removing disallowed characters
// and appending a hash if the name is too long or contains illegal characters.
// This function calls SanitizeID with alwaysAppendHash set to false.
func SanitizeToolName(name string) (string, error) {
	return SanitizeID([]string{name}, false, maxSanitizedPrefixLength, hashLength)
}

const (
	maxSanitizedPrefixLength = 53
	hashLength               = 8
	maxGeneratedIDLength     = maxSanitizedPrefixLength + 1 + hashLength
)

var (
	// nonWordChars is a regular expression that matches any character that is not a word character.
	nonWordChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	// disallowedIDChars is a regular expression that matches any character that is
	// not a valid character in an operation ID.
	disallowedIDChars = regexp.MustCompile(`[^a-zA-Z0-9-._~:/?#\[\]@!$&'()*+,;=]+`)
)

// GenerateUUID creates a new version 4 UUID and returns it as a string.
func GenerateUUID() string {
	return uuid.New().String()
}

// ParseToolName deconstructs a fully qualified tool name into its service key
// and bare tool name components. It splits the name using the standard
// separator.
//
// toolName is the fully qualified tool name.
// It returns the service key, the bare tool name, and an error if parsing fails
// (though the current implementation does not return an error).
func ParseToolName(toolName string) (service, bareToolName string, err error) {
	parts := strings.SplitN(toolName, consts.ToolNameServiceSeparator, 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", toolName, nil
}

// SanitizeOperationID cleans an input string to make it suitable for use as an
// operation ID. It replaces any sequence of disallowed characters with a short
// hexadecimal hash of that sequence, ensuring uniqueness while preserving as
// much of the original string as possible.
//
// input is the string to be sanitized.
// It returns the sanitized string.
func SanitizeOperationID(input string) string {
	if !disallowedIDChars.MatchString(input) {
		return input
	}

	// Use ReplaceAllStringFunc to generate a unique hash for each match
	sanitized := disallowedIDChars.ReplaceAllStringFunc(input, func(s string) string {
		h := sha256.New()
		h.Write([]byte(s))
		hash := hex.EncodeToString(h.Sum(nil))[:6]
		return fmt.Sprintf("_%s_", hash)
	})

	return sanitized
}

// GetDockerCommand returns the command and base arguments for running Docker,
// respecting the USE_SUDO_FOR_DOCKER environment variable.
func GetDockerCommand() (string, []string) {
	const dockerCmd = "docker"
	if os.Getenv("USE_SUDO_FOR_DOCKER") == "true" {
		return "sudo", []string{dockerCmd}
	}
	return dockerCmd, []string{}
}

// ReplaceURLPath replaces placeholders in a URL path with values from a params map.
func ReplaceURLPath(urlPath string, params map[string]interface{}) string {
	for k, v := range params {
		urlPath = strings.ReplaceAll(urlPath, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return urlPath
}
