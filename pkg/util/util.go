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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/mcpxy/core/pkg/consts"
)

var (
	validIDPattern    = regexp.MustCompile(`^[\w/-]+$`)
	disallowedIDChars = regexp.MustCompile(`[^a-zA-Z0-9-._~:/?#\[\]@!$&'()*+,;=]+`)
)

// GenerateToolID creates a fully qualified tool ID by combining a service key
// and a tool name. If the service key is empty, the tool name is returned as is.
//
// serviceKey is the unique identifier for the service.
// toolName is the name of the tool within the service.
// It returns the fully qualified tool ID or an error if the tool name is
// invalid.
func GenerateToolID(serviceKey, toolName string) (string, error) {
	if toolName == "" {
		return "", fmt.Errorf("tool name cannot be empty")
	}
	if !validIDPattern.MatchString(toolName) {
		return "", fmt.Errorf("tool name must match %q", validIDPattern.String())
	}

	if serviceKey != "" {
		// If the tool name already starts with the service key and the separator,
		// it is considered fully qualified and returned as is.
		if strings.HasPrefix(toolName, serviceKey+consts.ToolNameServiceSeparator) {
			return toolName, nil
		}
		// Otherwise, prepend the service key and separator.
		return serviceKey + consts.ToolNameServiceSeparator + toolName, nil
	}

	return toolName, nil
}

// GenerateServiceKey validates and returns a service key. In the current
// implementation, it simply validates that the service ID is not empty and
// conforms to the valid ID pattern.
//
// serviceID is the identifier for the service.
// It returns the validated service key or an error if the service ID is invalid.
func GenerateServiceKey(serviceID string) (string, error) {
	if serviceID == "" {
		return "", fmt.Errorf("service ID cannot be empty")
	}
	if !validIDPattern.MatchString(serviceID) {
		return "", fmt.Errorf("service ID must match %q", validIDPattern.String())
	}

	return serviceID, nil
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

	// Find all disallowed character sequences
	matches := disallowedIDChars.FindAllString(input, -1)
	if matches == nil {
		return input
	}

	// Join all disallowed sequences and compute a single hash
	fullInvalidSequence := strings.Join(matches, "")
	h := sha1.New()
	h.Write([]byte(fullInvalidSequence))
	hash := hex.EncodeToString(h.Sum(nil))[:6]

	// Replace every occurrence of a disallowed sequence with the same hash
	sanitized := disallowedIDChars.ReplaceAllString(input, fmt.Sprintf("_%s_", hash))

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
