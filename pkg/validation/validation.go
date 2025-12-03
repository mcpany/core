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

package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// IsValidBindAddress checks if a given string is a valid bind address.
// A valid bind address is in the format "host:port".
func IsValidBindAddress(s string) error {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if port == "" {
		return fmt.Errorf("port is required")
	}
	return nil
}

// IsSecurePath checks if a given file path is secure.
//
// A path is considered secure if it is relative (not absolute) and does not
// contain any path traversal sequences (e.g., "../") that could resolve to a
// location outside of the intended directory. This function is crucial for
// preventing directory traversal attacks.
//
// The function performs the following checks:
// 1. Verifies that the path is not absolute using `filepath.IsAbs`.
// 2. Cleans the path using `filepath.Clean` to resolve any ".." sequences.
// 3. Ensures that the cleaned path does not start with "../", which would
//    indicate a path traversal attempt.
//
// path is the file path to be validated.
//
// It returns an error if the path is found to be insecure (either absolute or
// containing directory traversal sequences), and nil otherwise.
func IsSecurePath(path string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed: %s", path)
	}
	cleanedPath := filepath.Clean(path)
	if strings.HasPrefix(cleanedPath, "..") {
		return fmt.Errorf("path resolves to a location outside of the working directory: %s", cleanedPath)
	}
	return nil
}

// IsValidURL checks if a given string is a valid URL. This function performs
// several checks, including for length, whitespace, the presence of a scheme,
// and host, considering special cases for schemes like "unix" or "mailto" that
// do not require a host.
//
// s is the string to be validated.
// It returns true if the string is a valid URL, and false otherwise.
func IsValidURL(s string) bool {
	if len(s) > 2048 || strings.TrimSpace(s) != s || strings.Contains(s, " ") {
		return false
	}

	u, err := url.Parse(s)
	if err != nil {
		return false
	}

	if u.Scheme == "" {
		return false
	}
	// If a host is NOT present, the scheme must be one that allows an opaque part.
	if u.Host == "" {
		// These schemes are allowed to not have a host component.
		allowedOpaqueSchemes := map[string]bool{
			"dns":         true,
			"unix":        true,
			"passthrough": true,
			"mailto":      true,
			"data":        true,
		}
		if !allowedOpaqueSchemes[u.Scheme] {
			return false
		}
		// Also ensure there is *something* after the scheme.
		if u.Opaque == "" && u.Path == "" {
			return false
		}
	} else {
		// If a host is present, it must not be only a port.
		if strings.HasPrefix(u.Host, ":") {
			return false
		}
	}

	return true
}

// ValidateHTTPServiceDefinition checks the validity of an HttpCallDefinition.
// It ensures that the endpoint path is specified and correctly formatted, and
// that a valid HTTP method is set.
//
// def is the HttpCallDefinition to be validated.
// It returns an error if the definition is invalid, and nil otherwise.
func ValidateHTTPServiceDefinition(def *configv1.HttpCallDefinition) error {
	if def == nil {
		return fmt.Errorf("http call definition cannot be nil")
	}
	if strings.TrimSpace(def.GetEndpointPath()) == "" {
		return fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(def.GetEndpointPath(), "/") {
		return fmt.Errorf("path must start with a '/'")
	}
	u, err := url.Parse(def.GetEndpointPath())
	if err != nil {
		return fmt.Errorf("path contains invalid characters")
	}
	if u.RawQuery != "" {
		return fmt.Errorf("path must not contain query parameters")
	}
	if def.GetMethod() == configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED {
		return fmt.Errorf("method is required")
	}
	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}
