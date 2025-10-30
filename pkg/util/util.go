/*
 * Copyright 2025 Author(s) of MCP-XY
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

package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/mcpxy/core/pkg/consts"
)

const (
	maxSanitizedPrefixLength = 53
	hashLength               = 8
	maxGeneratedIDLength     = maxSanitizedPrefixLength + 1 + hashLength
)

var (
	// nonWordChars is a regular expression that matches any character that is not a word character.
	nonWordChars = regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	// disallowedIDChars is a regular expression that matches any character that is
	// not a valid character in an operation ID.
	disallowedIDChars = regexp.MustCompile(`[^a-zA-Z0-9-._~:/?#\[\]@!$&'()*+,;=]`)
)

func sanitizeID(ids []string, alwaysAppendHash bool, maxSanitizedPrefixLength, hashLength int) (string, error) {
	var sanitizedIDs []string
	for _, id := range ids {
		if id == "" {
			return "", fmt.Errorf("id cannot be empty")
		}

		sanitizedID := nonWordChars.ReplaceAllString(id, "")
		appendHash := alwaysAppendHash || len(sanitizedID) > maxSanitizedPrefixLength || nonWordChars.MatchString(id)

		if len(sanitizedID) > maxSanitizedPrefixLength {
			sanitizedID = sanitizedID[:maxSanitizedPrefixLength]
		}

		if appendHash {
			h := sha256.New()
			h.Write([]byte(id))
			hash := hex.EncodeToString(h.Sum(nil))[:hashLength]
			sanitizedID = fmt.Sprintf("%s_%s", sanitizedID, hash)
		}
		sanitizedIDs = append(sanitizedIDs, sanitizedID)
	}
	return strings.Join(sanitizedIDs, "."), nil
}

// SanitizeServiceName sanitizes a service name.
func SanitizeServiceName(name string) (string, error) {
	return sanitizeID([]string{name}, false, maxSanitizedPrefixLength, hashLength)
}

// SanitizeToolName sanitizes a tool name.
func SanitizeToolName(serviceName, toolName string) (string, error) {
	return sanitizeID([]string{serviceName, toolName}, false, maxSanitizedPrefixLength, hashLength)
}

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
	if os.Getenv("USE_SUDO_FOR_DOCKER") == "true" {
		return "sudo", []string{"docker"}
	}
	return "docker", []string{}
}
