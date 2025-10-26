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

package auth

import (
	"context"
	"fmt"
	"net/http"

	configv1 "github.com/mcpxy/core/proto/config/v1"
)

// Authenticator defines the interface for authentication strategies. Each
// implementation is responsible for authenticating an incoming request and
// returning a context, which may be modified to include authentication-related
// information.
type Authenticator interface {
	// Authenticate validates the credentials in an HTTP request. It returns a
	// context, which may be enriched with authentication details, and an error if
	// the authentication fails.
	Authenticate(ctx context.Context, r *http.Request) (context.Context, error)
}

// APIKeyAuthenticator provides an authentication mechanism based on a static
// API key. It implements the `Authenticator` interface and checks for the
// presence of a specific header, validating its value against a configured key.
type APIKeyAuthenticator struct {
	HeaderName  string
	HeaderValue string
}

// NewAPIKeyAuthenticator creates a new APIKeyAuthenticator from the provided
// configuration. It returns `nil` if the configuration is invalid (e.g., if
// the header name or key value is missing).
//
// Parameters:
//   - config: The API key authentication settings, including the header
//     parameter name and the key value.
//
// Returns a new instance of APIKeyAuthenticator or `nil` if the configuration
// is invalid.
func NewAPIKeyAuthenticator(config *configv1.APIKeyAuth) *APIKeyAuthenticator {
	if config == nil || config.GetParamName() == "" || config.GetKeyValue() == "" {
		return nil
	}
	return &APIKeyAuthenticator{
		HeaderName:  config.GetParamName(),
		HeaderValue: config.GetKeyValue(),
	}
}

// Authenticate verifies the API key in the request headers. It checks if the
// header specified by `HeaderName` matches the expected `HeaderValue`.
//
// If the API key is valid, the original context is returned with no error. If
// the key is invalid or missing, an "unauthorized" error is returned.
//
// Parameters:
//   - ctx: The request context, which is returned unmodified on success.
//   - r: The HTTP request to authenticate.
//
// Returns the original context and `nil` on success, or an error on failure.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	if r.Header.Get(a.HeaderName) == a.HeaderValue {
		return ctx, nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// AuthManager oversees the authentication process for the server. It maintains a
// registry of authenticators, each associated with a specific service ID, and
// delegates the authentication of requests to the appropriate authenticator.
// This allows for different authentication strategies to be used for different
// services.
type AuthManager struct {
	authenticators map[string]Authenticator
}

// NewAuthManager creates and initializes a new AuthManager with an empty
// authenticator registry. This manager can then be used to register and manage
// authenticators for various services.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		authenticators: make(map[string]Authenticator),
	}
}

// AddAuthenticator registers an authenticator for a given service ID. If an
// authenticator is already registered for the same service ID, it will be
// overwritten.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//   - authenticator: The authenticator to be associated with the service.
func (am *AuthManager) AddAuthenticator(serviceID string, authenticator Authenticator) {
	am.authenticators[serviceID] = authenticator
}

// Authenticate authenticates a request for a specific service. It looks up the
// authenticator registered for the given service ID and, if found, uses it to
// validate the request.
//
// If no authenticator is found for the service, the request is allowed to
// proceed without authentication.
//
// Parameters:
//   - ctx: The request context.
//   - serviceID: The identifier of the service being accessed.
//   - r: The HTTP request to authenticate.
//
// Returns a potentially modified context on success, or an error if
// authentication fails.
func (am *AuthManager) Authenticate(ctx context.Context, serviceID string, r *http.Request) (context.Context, error) {
	if authenticator, ok := am.authenticators[serviceID]; ok {
		return authenticator.Authenticate(ctx, r)
	}
	// If no authenticator is configured for the service, we'll allow the request to proceed.
	return ctx, nil
}

// GetAuthenticator retrieves the authenticator registered for a specific
// service.
//
// Parameters:
//   - serviceID: The identifier of the service.
//
// Returns the authenticator and a boolean indicating whether an authenticator
// was found.
func (am *AuthManager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	authenticator, ok := am.authenticators[serviceID]
	return authenticator, ok
}

// AddOAuth2Authenticator creates and registers a new OAuth2Authenticator for a
// given service ID. It initializes the authenticator using the provided OAuth2
// configuration.
//
// This is a convenience method that simplifies the process of setting up OAuth2
// authentication for a service.
//
// Parameters:
//   - ctx: The context for initializing the OIDC provider.
//   - serviceID: The unique identifier for the service.
//   - config: The OAuth2 configuration for the authenticator.
//
// Returns an error if the authenticator cannot be created.
func (am *AuthManager) AddOAuth2Authenticator(ctx context.Context, serviceID string, config *OAuth2Config) error {
	if config == nil {
		return nil
	}
	authenticator, err := NewOAuth2Authenticator(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create OAuth2 authenticator for service %s: %w", serviceID, err)
	}
	am.AddAuthenticator(serviceID, authenticator)
	return nil
}
