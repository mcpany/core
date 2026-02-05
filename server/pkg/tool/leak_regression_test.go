// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"net/url"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_RedactURL_LeaksUserInfo(t *testing.T) {
	// Setup
	// URL with user info and a secret query parameter
	rawURL := "https://user:password@example.com/api?secret_param=MY_SECRET"
	u, err := url.Parse(rawURL)
	require.NoError(t, err)

	// Construct parameter mapping using setters/builders because fields are private
	schema := &configv1.ParameterSchema{}
	schema.SetName("secret_param")

	secret := &configv1.SecretValue{}
	// SecretValue is a oneof. We just need to set it to something non-nil.

	mapping := &configv1.HttpParameterMapping{}
	mapping.SetSchema(schema)
	mapping.SetSecret(secret)

	// Create a dummy HTTPTool
	tool := &HTTPTool{
		parameters: []*configv1.HttpParameterMapping{
			mapping,
		},
	}

	// Test redactURL
	redacted := tool.redactURL(u)

	// Verification
	// The password should be replaced with [REDACTED].
	// Since it's part of the URL, it might be encoded.
	// The redactURL method uses url.UserPassword which does NOT encode [ or ] if they are not reserved in user info?
	// Wait, url.UserPassword("user", "[REDACTED]") -> user:[REDACTED]
	// But url.String() encodes it if necessary.
	// [ and ] are reserved in some parts but maybe not in password?
	// Actually, url.String() encoding depends on the Go version and implementation.
	// In Go 1.23+, it might behave differently.
	// Let's check for both encoded and unencoded just in case, or verify what we saw in output: %5BREDACTED%5D

	assert.NotContains(t, redacted, "password", "Password should NOT be present in redacted URL")

	// Check for redacted password. From previous run we know it is encoded.
	assert.Contains(t, redacted, "%5BREDACTED%5D", "Password should be redacted (and encoded as needed)")

	// Also verify that the query parameter WAS redacted
	assert.Contains(t, redacted, "secret_param=%5BREDACTED%5D", "Secret query parameter should be redacted")
}
