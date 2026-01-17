// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package validation provides validation utilities for config files and other inputs.
package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// IsValidBindAddress checks if a given string is a valid bind address.
// A valid bind address is in the format "host:port".
func IsValidBindAddress(s string) error {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		// If the error is due to missing port in address (which happens when no colon is present),
		// we check if it is a valid port-only string by prepending ":".
		if !strings.Contains(s, ":") {
			// Ensure it is a numeric port to distinguish from missing-port hostnames like "localhost"
			if p, err := strconv.Atoi(s); err == nil {
				if p < 0 || p > 65535 {
					return fmt.Errorf("port must be between 0 and 65535")
				}
				_, port, err = net.SplitHostPort(":" + s)
				if err == nil && port != "" {
					return nil
				}
			}
		}
		return err
	}
	if port == "" {
		return fmt.Errorf("port is required")
	}
	// Check if port is numeric and within range
	p, err := strconv.Atoi(port)
	if err != nil {
		// Non-numeric ports (service names) are allowed.
		return nil //nolint:nilerr // Intentional: we allow non-numeric ports
	}
	if p < 0 || p > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	return nil
}

// IsSecurePath checks if a given file path is secure and does not contain any
// path traversal sequences ("../" or "..\\"). This function is crucial for
// preventing directory traversal attacks, where a malicious actor could
// otherwise access or manipulate files outside of the intended directory.
//
// path is the file path to be validated.
// It returns an error if the path is found to be insecure, and nil otherwise.
// IsSecurePath checks if a given file path is secure.
// It is a variable to allow mocking in tests.
var IsSecurePath = util.IsSecurePath

// SetAllowedPaths checks if a given file path is relative and does not contain any
// path traversal sequences ("../").
func SetAllowedPaths(paths []string) {
	util.SetAllowedPaths(paths)
}

// IsAllowedPath checks if a given file path is allowed (inside CWD or AllowedPaths)
// and does not contain any path traversal sequences ("../").
// It is a variable to allow mocking in tests.
var IsAllowedPath = util.IsAllowedPath

// allowedOpaqueSchemes are schemes that are allowed to not have a host component.
var allowedOpaqueSchemes = map[string]bool{
	"dns":         true,
	"unix":        true,
	"passthrough": true,
	"mailto":      true,
	"data":        true,
	"file":        true,
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
		if !allowedOpaqueSchemes[u.Scheme] {
			return false
		}
		// Also ensure there is *something* after the scheme.
		if u.Opaque == "" && u.Path == "" {
			return false
		}
	} else if strings.HasPrefix(u.Host, ":") { // If a host is present, it must not be only a port.
		return false
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
var FileExists = func(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}
