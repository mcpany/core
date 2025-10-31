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
	"net/url"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

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
	if def.GetEndpointPath() == "" {
		return fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(def.GetEndpointPath(), "/") {
		return fmt.Errorf("path must start with a '/'")
	}
	if def.GetMethod() == configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED {
		return fmt.Errorf("method is required")
	}
	return nil
}
