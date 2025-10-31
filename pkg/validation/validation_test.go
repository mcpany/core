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
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		name     string
		rawURL   string
		expected bool
	}{
		// Valid URLs
		{"valid http", "http://example.com", true},
		{"valid https", "https://example.com/path?query=value#fragment", true},
		{"valid ftp", "ftp://user:pass@example.com/resource", true},
		{"valid localhost http", "http://localhost:8080", true},
		{"valid ip address http", "http://127.0.0.1/test", true},
		{"valid dns scheme simple", "dns:example.com", true},
		{"valid dns scheme full", "dns:///resolver.example.com/example.com", true},   // Host is "resolver.example.com"
		{"valid unix scheme with path", "unix:/tmp/socket.sock", true},               // No host, but path is present
		{"valid unix scheme with slashes and path", "unix:///tmp/socket.sock", true}, // No host, path is present
		{"valid passthrough scheme with path", "passthrough:///service-name", true},  // No host, path is present
		{"valid passthrough scheme opaque", "passthrough:service-name", true},        // No host, opaque is present
		{"valid mailto", "mailto:user@example.com", true},
		{"valid data url", "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==", true},

		// Invalid URLs
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"just scheme http", "http:", false},
		{"just scheme https with slashes", "https://", false}, // common web scheme, requires host
		{"missing scheme", "example.com", false},
		{"missing host for http", "http://", false}, // common web scheme, requires host
		{"http scheme with empty authority", "http:///", false},
		{"http scheme with empty authority and path", "http:///path", false},
		{"http scheme with port but no host", "http://:8080", false},
		{"scheme only custom", "customscheme:", false}, // No host, no opaque, no path
		{"url with internal spaces", "http://example.com/path with spaces", false},
		{"url with leading space", " http://example.com", false},
		{"url with trailing space", "http://example.com ", false},
		{"url with internal and surrounding spaces", " http://example.com/path with spaces ", false},
		// Base length of "http://example.com/" is 19. 2048 - 19 = 2029.
		{"very long url (just at limit)", "http://example.com/" + strings.Repeat("a", 2029), true},           // Total 2048
		{"excessively long url (just over limit)", "http://example.com/" + strings.Repeat("a", 2030), false}, // Total 2049
		{"no scheme but has slashes", "//example.com/path", false},
		{"invalid scheme chars", "ht!tp://example.com", false},
		{"dns scheme malformed (empty opaque/path)", "dns:", false},

		// Specific gRPC target cases
		{"grpc target no scheme", "localhost:50051", false},                       // Invalid: No scheme
		{"grpc target with scheme", "grpc://localhost:50051", true},               // Valid: Has scheme and host
		{"dns target for grpc", "dns:resolver.example.com/service", true},         // Valid: dns scheme with opaque part
		{"dns target for grpc full", "dns:///resolver.example.com/service", true}, // Valid: dns scheme with host and path
		{"unix target for grpc", "unix:/var/run/service.sock", true},              // Valid: unix scheme with path
		{"unix target for grpc full", "unix:///var/run/service.sock", true},       // Valid: unix scheme with path
		{"passthrough target for grpc", "passthrough:///my_service", true},        // Valid: passthrough scheme with path
		{"passthrough target opaque for grpc", "passthrough:my_service", true},    // Valid: passthrough scheme with opaque part
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsValidURL(tc.rawURL)
			if actual != tc.expected {
				t.Errorf("IsValidURL(%q) = %v; want %v", tc.rawURL, actual, tc.expected)
			}
		})
	}
}

func TestValidateHTTPServiceDefinition(t *testing.T) {
	methodPost := configv1.HttpCallDefinition_HTTP_METHOD_POST
	methodGet := configv1.HttpCallDefinition_HTTP_METHOD_GET
	methodUnspecified := configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED

	testCases := []struct {
		name          string
		def           *configv1.HttpCallDefinition
		expectedError string
	}{
		{
			name: "valid definition",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("/v1/users"),
				Method:       &methodPost,
			}.Build(),
			expectedError: "",
		},
		{
			name:          "nil definition",
			def:           nil,
			expectedError: "http call definition cannot be nil",
		},
		{
			name: "missing path",
			def: configv1.HttpCallDefinition_builder{
				Method: &methodGet,
			}.Build(),
			expectedError: "path is required",
		},
		{
			name: "path does not start with slash",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("v1/users"),
				Method:       &methodGet,
			}.Build(),
			expectedError: "path must start with a '/'",
		},
		{
			name: "unspecified method",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("/v1/users"),
				Method:       &methodUnspecified,
			}.Build(),
			expectedError: "method is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHTTPServiceDefinition(tc.def)
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
