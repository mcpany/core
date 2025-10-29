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
	maxGeneratedIDLength     = 62
)

var (
	// nonWordChars is a regular expression that matches any character that is not a word character.
	nonWordChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	// wordChars is a regular expression that matches only word characters.
	wordChars = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	// disallowedIDChars is a regular expression that matches any character that is
	// not a valid character in an operation ID.
	disallowedIDChars = regexp.MustCompile(`[^a-zA-Z0-9-._~:/?#\[\]@!$&'()*+,;=]`)
)

// GenerateID creates a unique and consistently formatted identifier from a given set of parts.
// For each part, if it's considered valid (matches `wordChars` and is within `maxGeneratedIDLength`),
// it's used as is. Otherwise, it is sanitized and hashed. The final ID is created by joining
// the processed parts with a ".".
//
// The sanitization format is <sanitized_prefix>_<hash>, with a maximum length of 62 characters.
// The <sanitized_prefix> is the first 53 characters of the original part, with all non-word characters removed.
// The <hash> is the first 8 characters of the SHA256 hash of the original part.
func GenerateID(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("at least one part must be provided")
	}

	processedParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("name parts cannot be empty")
		}

		// Check if the part is valid and within the length limit
		if wordChars.MatchString(part) && len(part) <= maxGeneratedIDLength {
			processedParts = append(processedParts, part)
			continue
		}

		// If not valid, sanitize and hash it
		h := sha256.New()
		h.Write([]byte(part))
		hash := hex.EncodeToString(h.Sum(nil))[:hashLength]

		sanitizedPrefix := nonWordChars.ReplaceAllString(part, "")
		if len(sanitizedPrefix) > maxSanitizedPrefixLength {
			sanitizedPrefix = sanitizedPrefix[:maxSanitizedPrefixLength]
		}

		processedPart := fmt.Sprintf("%s_%s", sanitizedPrefix, hash)
		processedParts = append(processedParts, processedPart)
	}

	return strings.Join(processedParts, consts.ToolNameServiceSeparator), nil
}

// GenerateToolName generates a unique and consistently formatted name for a tool.
// It uses the GenerateID function to ensure the name is valid and unique.
//
// toolName is the name of the tool.
// It returns the generated tool name or an error if the tool name is empty.
func GenerateToolName(toolName string) (string, error) {
	return GenerateID(toolName)
}

// GenerateUUID creates a new version 4 UUID and returns it as a string.
func GenerateUUID() string {
	return uuid.New().String()
}

// ParseToolName deconstructs a fully qualified tool name into its service ID
// and bare tool name components. It splits the name using the standard
// separator.
//
// toolName is the fully qualified tool name.
// It returns the service ID, the bare tool name, and an error if parsing fails
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
