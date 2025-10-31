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

package auth

// OAuth2Config holds the configuration for OAuth2 authentication. It is used to
// configure the OAuth2Authenticator with the necessary parameters to validate
// JWTs against an OIDC provider.
type OAuth2Config struct {
	// IssuerURL is the URL of the OIDC provider's issuer. This is used to
	// fetch the provider's public keys for token validation.
	IssuerURL string
	// Audience is the intended audience of the JWT. The authenticator will
	// verify that the token's 'aud' claim matches this value.
	Audience string
}
