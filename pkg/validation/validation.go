// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package validation provides validation utilities for config files and other inputs.
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

// IsSecurePath checks if a given file path is secure and does not contain any
// path traversal sequences ("../" or "..\\"). This function is crucial for
// preventing directory traversal attacks, where a malicious actor could
// otherwise access or manipulate files outside of the intended directory.
//
// path is the file path to be validated.
// It returns an error if the path is found to be insecure, and nil otherwise.
// IsSecurePath checks if a given file path is secure.
// It is a variable to allow mocking in tests.
var IsSecurePath = func(path string) error {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("path contains '..', which is not allowed")
		}
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
func FileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}
