// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package auth provides authentication and authorization functionality.
package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"sync"
	// UserContextKey is the context key for the user ID.
	//
	// Summary: Defines UserContextKey.
// ContextWithAPIKey returns a new context with the API Key embedded.
//
// Summary: Embeds an API key into the context.
// APIKeyFromContext returns the API Key from the context if present.
//
// Summary: Retrieves the API key from the context.
//
// ContextWithUser returns a new context with the user ID embedded.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Summary: Embeds a user ID into the context.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// UserFromContext returns the user ID from the context if present.
//
// Summary: Retrieves the user ID from the context.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// ContextWithProfileID returns a new context with the profile ID embedded.
//
// Summary: Embeds a profile ID into the context.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// ProfileIDFromContext returns the profile ID from the context if present.
//
// Summary: Retrieves the profile ID from the context.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - string: The profile ID.
//   - bool: True if found.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
// Authenticator defines the interface for authentication mechanisms.
//
// Summary: Interface for authenticating HTTP requests.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Authenticator interface {
	// Authenticate checks if a request is authenticated and returns the updated context.
	//
	// Summary: Authenticates a request.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - r: *http.Request. The HTTP request.
	//
	// Returns:
// NewAPIKeyAuthenticator creates a new APIKeyAuthenticator instance.
//
// Summary: Initializes an APIKeyAuthenticator.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Parameters:
//   - config: *configv1.APIKeyAuth. The configuration settings.
//
// Returns:
// Authenticate verifies the API key in the request.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Validates the API key from header, query, or cookie.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Context with API key if valid.
//   - error: Error if unauthorized.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	var receivedKey string
	switch a.In {
	case configv1.APIKeyAuth_HEADER:
		receivedKey = r.Header.Get(a.ParamName)
	case configv1.APIKeyAuth_QUERY:
		receivedKey = r.URL.Query().Get(a.ParamName)
	case configv1.APIKeyAuth_COOKIE:
		cookie, err := r.Cookie(a.ParamName)
		if err == nil {
			receivedKey = cookie.Value
		}
	default:
		receivedKey = r.Header.Get(a.ParamName)
	}
// NewBasicAuthenticator creates a new BasicAuthenticator instance.
//
// Summary: Initializes a BasicAuthenticator.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - config: *configv1.BasicAuth. The configuration settings.
//
// Authenticate validates the basic auth credentials.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Validates username and password hash.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Authenticated context.
//   - error: Error if unauthorized.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (a *BasicAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	user, password, ok := r.BasicAuth()
	if !ok {
		return ctx, fmt.Errorf("unauthorized")
	}

	usernameMatch := true
	if a.Username != "" {
		if subtle.ConstantTimeCompare([]byte(user), []byte(a.Username)) != 1 {
			usernameMatch = false
		}
	}

	// Always check password to avoid timing attacks that could reveal if the username is correct
	passwordMatch := passhash.CheckPassword(password, a.PasswordHash)
// NewTrustedHeaderAuthenticator creates a new TrustedHeaderAuthenticator instance.
//
// Summary: Initializes a TrustedHeaderAuthenticator.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - config: *configv1.TrustedHeaderAuth. The configuration settings.
//
// Authenticate validates the trusted header.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Checks for the trusted header.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Authenticated context.
//   - error: Error if unauthorized.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (a *TrustedHeaderAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	val := r.Header.Get(a.HeaderName)
	if val == "" {
		return ctx, fmt.Errorf("unauthorized")
	}
	// If HeaderValue is set, it must match.
	if a.HeaderValue != "" {
		if subtle.ConstantTimeCompare([]byte(val), []byte(a.HeaderValue)) != 1 {
			return ctx, fmt.Errorf("unauthorized")
		}
// Manager oversees the authentication process for the server.
//
// Summary: Manages authentication strategies and user sessions.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Manager struct {
	authenticators *xsync.Map[string, Authenticator]
	apiKey         string

	// usersMu protects users map to allow atomic updates (hot-swap).
	usersMu sync.RWMutex
// NewManager creates and initializes a new Manager with an empty authenticator registry.
//
// Summary: Initializes a new Authentication Manager.
// SetUsers updates the list of active users.
//
// Summary: Sets the configured users.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - users: []*configv1.User. The list of users.
//
// SetStorage sets the storage backend for the manager.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Configures the storage backend.
//
// GetUser retrieves a user configuration by their ID.
//
// Summary: Looks up a user by ID.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - id: string. The user ID.
// AddAuthenticator registers an authenticator for a given service ID.
//
// Summary: Registers an authenticator for a service.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - serviceID: string. The service ID.
//   - authenticator: Authenticator. The authenticator instance.
//
// Authenticate authenticates a request for a specific service.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Authenticates a request, checking service-specific or global rules.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - ctx: context.Context. The request context.
//   - serviceID: string. The service ID.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: Error if unauthorized.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (am *Manager) Authenticate(ctx context.Context, serviceID string, r *http.Request) (context.Context, error) {
	if am.apiKey != "" {
		receivedKey := r.Header.Get("X-API-Key")
		if receivedKey == "" {
			receivedKey = r.URL.Query().Get("api_key")
		}

		if receivedKey == "" {
			return ctx, fmt.Errorf("unauthorized")
		}
		if subtle.ConstantTimeCompare([]byte(receivedKey), []byte(am.apiKey)) != 1 {
			return ctx, fmt.Errorf("unauthorized")
		}
		ctx = ContextWithAPIKey(ctx, receivedKey)
	}

	if authenticator, ok := am.authenticators.Load(serviceID); ok {
		return authenticator.Authenticate(ctx, r)
	}
	// If no authenticator is configured for the service:
	// If we authenticated via Global API Key, we allow it.
	// If not found, check global keys...
	if am.apiKey != "" {
		// If we authenticated via Global API Key, we allow it.
		// NOTE: logic was: if apiKey configured, and we reached here (meaning no service authenticator), allow.
		// But wait, if apiKey was provided, we updated ctx.
		// If apiKey was NOT provided, we still fall through?
		// Authenticate logic above:
		// if am.apiKey != "" { check header }
		// if header valid, ctx updated.
		// If header missing, returns error "unauthorized" IF apiKey is required?
		// Logic at 373: if receivedKey == "" return error.
		// So if apiKey is configured, we MUST provide it?
		// Check lines 365-378.
		// Yes, if am.apiKey != "", we ENFORCE it.
		// So if we are here, we passed API key check (if configured).
		// So we can return nil (allow).
		return ctx, nil
// GetAuthenticator retrieves the authenticator registered for a specific service.
//
// Summary: Looks up an authenticator by service ID.
//
// RemoveAuthenticator removes the authenticator for a given service ID.
// AddOAuth2Authenticator creates and registers a new OAuth2Authenticator for a given service ID.
//
// Summary: Helper to add an OAuth2 authenticator.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - ctx: context.Context. Context for initialization.
//   - serviceID: string. The service ID.
//   - config: *OAuth2Config. The OAuth2 configuration.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: Error if creation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (am *Manager) AddOAuth2Authenticator(ctx context.Context, serviceID string, config *OAuth2Config) error {
	if config == nil {
		return nil
	}
// ValidateAuthentication validates the authentication request against the provided configuration.
//
// Summary: Validates a request against a specific auth configuration.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - config: *configv1.Authentication. The authentication configuration.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - error: Error if validation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func ValidateAuthentication(ctx context.Context, config *configv1.Authentication, r *http.Request) error {
	if config == nil {
		return nil // No auth configured implies allowed
	}

	switch config.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		authenticator := NewAPIKeyAuthenticator(config.GetApiKey())
		if authenticator == nil {
			return fmt.Errorf("invalid API key configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case configv1.Authentication_Oauth2_case:
		cfg := config.GetOauth2()
		if cfg.GetIssuerUrl() == "" {
			return fmt.Errorf("invalid OAuth2 configuration: missing issuer_url")
		}
		cacheKey := cfg.GetIssuerUrl() + "|" + cfg.GetAudience()

		authenticator, ok := oauthAuthenticatorCache.Load(cacheKey)
		if !ok {
			oConfig := &OAuth2Config{
				IssuerURL: cfg.GetIssuerUrl(),
				Audience:  cfg.GetAudience(),
			}
			// Use context.Background() with a timeout for authenticator initialization to avoid
			// binding the OIDC provider to a short-lived request context and prevent hanging.
			initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			newAuth, err := NewOAuth2Authenticator(initCtx, oConfig)
			if err != nil {
				return fmt.Errorf("failed to create oauth2 authenticator: %w", err)
			}
			// Race condition handling: check if someone else inserted it
			actual, loaded := oauthAuthenticatorCache.LoadOrStore(cacheKey, newAuth)
			if loaded {
				authenticator = actual
			} else {
				authenticator = newAuth
			}
		}

		_, err := authenticator.Authenticate(ctx, r)
		return err
	case configv1.Authentication_BasicAuth_case:
		authenticator := NewBasicAuthenticator(config.GetBasicAuth())
		if authenticator == nil {
			return fmt.Errorf("invalid Basic Auth configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case configv1.Authentication_TrustedHeader_case:
		authenticator := NewTrustedHeaderAuthenticator(config.GetTrustedHeader())
		if authenticator == nil {
			return fmt.Errorf("invalid Trusted Header configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case configv1.Authentication_Oidc_case:
		cfg := config.GetOidc()
		if cfg.GetIssuer() == "" {
			return fmt.Errorf("invalid OIDC configuration: missing issuer")
		}

		audiences := cfg.GetAudience()
		audStr := strings.Join(audiences, ",")
		cacheKey := cfg.GetIssuer() + "|" + audStr

		authenticator, ok := oauthAuthenticatorCache.Load(cacheKey)
		if !ok {
			oConfig := &OAuth2Config{
				IssuerURL: cfg.GetIssuer(),
				Audiences: audiences,
			}
			initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			newAuth, err := NewOAuth2Authenticator(initCtx, oConfig)
			if err != nil {
				return fmt.Errorf("failed to create oidc authenticator: %w", err)
			}
			actual, loaded := oauthAuthenticatorCache.LoadOrStore(cacheKey, newAuth)
			if loaded {
				authenticator = actual
			} else {
				authenticator = newAuth
			}
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	default:
		return fmt.Errorf("unsupported or missing authentication method")
	}
}

// checkBasicAuthWithUsers checks if the request has valid Basic Auth credentials
// matching any of the configured users.
func (am *Manager) checkBasicAuthWithUsers(ctx context.Context, r *http.Request) (context.Context, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return ctx, fmt.Errorf("no basic auth provided")
	}

	am.usersMu.RLock()
	defer am.usersMu.RUnlock()

	// Direct lookup if user ID matches username
	if user, ok := am.users[username]; ok {
		if basicAuth := user.GetAuthentication().GetBasicAuth(); basicAuth != nil {
			if passhash.CheckPassword(password, basicAuth.GetPasswordHash()) {
				ctx = ContextWithUser(ctx, user.GetId())
				if len(user.GetRoles()) > 0 {
					ctx = ContextWithRoles(ctx, user.GetRoles())
				}
				return ctx, nil
			}
		}
	}

	// Fallback: Iterate all users (in case username is not ID, but we assume ID==Username for now)
	return ctx, fmt.Errorf("invalid credentials")
}
//   - None.
// Side Effects:
//   - None.
// Errors:
// Returns:
//   - None.
// Parameters:
// ContextWithUser returns a new context with the user ID embedded.
//
// Summary: Retrieves the API key from the context.
//
// APIKeyFromContext returns the API Key from the context if present.
// Summary: Embeds an API key into the context.
//
// ContextWithAPIKey returns a new context with the API Key embedded.
	// Summary: Defines UserContextKey.
	//
	// UserContextKey is the context key for the user ID.
