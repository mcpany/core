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
// API key. It checks for the presence of a specific header and validates its
// value.
type APIKeyAuthenticator struct {
	HeaderName  string
	HeaderValue string
}

// NewAPIKeyAuthenticator creates a new APIKeyAuthenticator from the provided
// configuration. It returns nil if the configuration is invalid.
//
// config contains the API key authentication settings, including the header
// parameter name and the key value.
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
// header specified by HeaderName matches the expected HeaderValue.
//
// ctx is the request context, which is returned unmodified on success.
// r is the HTTP request to authenticate.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	if r.Header.Get(a.HeaderName) == a.HeaderValue {
		return ctx, nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// AuthManager oversees the authentication process for the server. It maintains a
// registry of authenticators, each associated with a specific service ID, and
// delegates the authentication of requests to the appropriate authenticator.
type AuthManager struct {
	authenticators map[string]Authenticator
}

// NewAuthManager creates and initializes a new AuthManager with an empty
// authenticator registry.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		authenticators: make(map[string]Authenticator),
	}
}

// AddAuthenticator registers an authenticator for a given service ID.
//
// serviceID is the unique identifier for the service.
// authenticator is the authenticator to be associated with the service.
func (am *AuthManager) AddAuthenticator(serviceID string, authenticator Authenticator) {
	am.authenticators[serviceID] = authenticator
}

// Authenticate authenticates a request for a specific service. If an
// authenticator is registered for the service, it will be used to validate the
// request. If no authenticator is found, the request is allowed to proceed.
//
// ctx is the request context.
// serviceID is the identifier of the service being accessed.
// r is the HTTP request to authenticate.
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
// serviceID is the identifier of the service.
// It returns the authenticator and a boolean indicating whether an
// authenticator was found.
func (am *AuthManager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	authenticator, ok := am.authenticators[serviceID]
	return authenticator, ok
}

// AddOAuth2Authenticator creates and registers a new OAuth2Authenticator for the
// given service ID. It initializes the authenticator using the provided OAuth2
// configuration.
//
// ctx is the context for initializing the OIDC provider.
// serviceID is the unique identifier for the service.
// config is the OAuth2 configuration for the authenticator.
//
// It returns an error if the authenticator cannot be created.
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
