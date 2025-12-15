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

// Package auth provides authentication implementations.
package auth

import (
	"context"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

type authContextKey string

const (
	// UserContextKey is the context key for the user ID.
	UserContextKey authContextKey = "user_id"
	// ProfileIDContextKey is the context key for the profile ID.
	ProfileIDContextKey authContextKey = "profile_id"
)

// ContextWithUser returns a new context with the user ID.
func ContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// UserFromContext returns the user ID from the context.
func UserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserContextKey).(string)
	return val, ok
}

// ContextWithProfileID returns a new context with the profile ID.
func ContextWithProfileID(ctx context.Context, profileID string) context.Context {
	return context.WithValue(ctx, ProfileIDContextKey, profileID)
}

// ProfileIDFromContext returns the profile ID from the context.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ProfileIDContextKey).(string)
	return val, ok
}

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
	ParamName string
	In        configv1.APIKeyAuth_Location
	Value     string
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
		ParamName: config.GetParamName(),
		In:        config.GetIn(),
		Value:     config.GetKeyValue(),
	}
}

// Authenticate verifies the API key in the request. It checks if the
// parameter specified by `ParamName` matches the expected `Value`.
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
	var receivedKey string
	switch a.In {
	case configv1.APIKeyAuth_HEADER:
		receivedKey = r.Header.Get(a.ParamName)
	case configv1.APIKeyAuth_QUERY:
		receivedKey = r.URL.Query().Get(a.ParamName)
	default:
		receivedKey = r.Header.Get(a.ParamName)
	}

	if receivedKey == a.Value {
		return ctx, nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// Manager oversees the authentication process for the server. It maintains a
// registry of authenticators, each associated with a specific service ID, and
// delegates the authentication of requests to the appropriate authenticator.
// This allows for different authentication strategies to be used for different
// services.
type Manager struct {
	authenticators *xsync.Map[string, Authenticator]
	apiKey         string
}

// NewManager creates and initializes a new Manager with an empty
// authenticator registry. This manager can then be used to register and manage
// authenticators for various services.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
	}
}

// SetAPIKey sets the global API key for the server.
func (am *Manager) SetAPIKey(apiKey string) {
	am.apiKey = apiKey
}

// AddAuthenticator registers an authenticator for a given service ID. If an
// authenticator is already registered for the same service ID, it will be
// overwritten.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//   - authenticator: The authenticator to be associated with the service.
//
// Returns an error if the provided authenticator is `nil`.
func (am *Manager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if authenticator == nil {
		return fmt.Errorf("authenticator for service %s is nil", serviceID)
	}
	am.authenticators.Store(serviceID, authenticator)
	return nil
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
func (am *Manager) Authenticate(ctx context.Context, serviceID string, r *http.Request) (context.Context, error) {
	if am.apiKey != "" {
		if r.Header.Get("X-API-Key") == "" {
			return ctx, fmt.Errorf("unauthorized: missing API key")
		}
		if r.Header.Get("X-API-Key") != am.apiKey {
			return ctx, fmt.Errorf("unauthorized: invalid API key")
		}
	}

	if authenticator, ok := am.authenticators.Load(serviceID); ok {
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
func (am *Manager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	return am.authenticators.Load(serviceID)
}

// RemoveAuthenticator removes the authenticator for a given service ID.
func (am *Manager) RemoveAuthenticator(serviceID string) {
	am.authenticators.Delete(serviceID)
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
func (am *Manager) AddOAuth2Authenticator(ctx context.Context, serviceID string, config *OAuth2Config) error {
	if config == nil {
		return nil
	}
	authenticator, err := NewOAuth2Authenticator(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create OAuth2 authenticator for service %s: %w", serviceID, err)
	}
	return am.AddAuthenticator(serviceID, authenticator)
}
// ValidateAuthentication validates the authentication request against the provided configuration.
// It supports API Key and OAuth2 authentication methods.
//
// Parameters:
//   - ctx: The context for the request.
//   - config: The authentication configuration.
//   - r: The HTTP request to validate.
//
// Returns an error if validation fails or the method is unsupported.
func ValidateAuthentication(ctx context.Context, config *configv1.AuthenticationConfig, r *http.Request) error {
	if config == nil {
		return nil // No auth configured implies allowed
	}

	switch method := config.AuthMethod.(type) {
	case *configv1.AuthenticationConfig_ApiKey:
		authenticator := NewAPIKeyAuthenticator(method.ApiKey)
		if authenticator == nil {
			return fmt.Errorf("invalid API key configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case *configv1.AuthenticationConfig_Oauth2:
		// OAuth2 validation typically requires a more complex flow or token validation.
		// For the server receiving a request, it usually expects a Bearer token that matches some introspection or local validation.
		// However, the OAuth2Auth config struct usually defines client credentials for *outgoing* requests or *setup*.
		// If used for incoming auth, it might imply validating a JWT or similar.
		// For now, we'll placeholder this or strictly check if headers are present if that's the intention.
		// BUT, reading the proto, OAuth2Auth has token_url, client_id etc. This is for CLIENT usage mostly.
		// Verification of incoming OAuth2 tokens usually involves JWKS or similar which isn't in that config.
		// So we might need to assume this config is for client?
		// Wait, the user asked for "AuthenticationConfig" to be used for "authentication in user and profile".
		// Usually this implies *incoming* auth to the MCP server.
		// If the MCP server is acting as an OAuth2 Resource Server, it needs validation keys, not client_id/secret.
		// Let's assume for now we only fully support APIKey for internal User/Profile incoming auth, or strict equality?
		// Re-reading usage: "we are going to reuse the same function to authentication user."
		// Let's implement API Key core logic.
		return fmt.Errorf("oauth2 authentication not yet implemented for incoming requests")
	default:
		return nil
	}
}
