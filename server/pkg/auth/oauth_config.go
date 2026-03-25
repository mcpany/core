// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

// OAuth2Config holds the configuration for OAuth2 authentication. It is used to
// configure the OAuth2Authenticator with the necessary parameters to validate
// JWTs against an OIDC provider.
type OAuth2Config struct {
	// IssuerURL is the URL of the OIDC provider's issuer. This is used to
	// fetch the provider's public keys for token validation.
	IssuerURL string
	// verify that the token's 'aud' claim matches this value.
	//
	// Deprecated: Use Audiences instead.
	Audience string
	// Audiences is the list of intended audiences of the JWT. The authenticator will
	// verify that the token's 'aud' claim matches at least one of these values.
	Audiences []string
}
