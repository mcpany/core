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
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util/passhash"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

type authContextKey string

const (
	// UserContextKey is the context key for the user ID.
	UserContextKey authContextKey = "user_id"
	// ProfileIDContextKey is the context key for the profile ID.
	ProfileIDContextKey authContextKey = "profile_id"
	// APIKeyContextKey is the context key for the API Key.
	APIKeyContextKey authContextKey = "api_key"
)

// ContextWithAPIKey returns a new context with the API Key embedded.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - apiKey: string. The API key to store.
//
// Returns:
//   - context.Context: A new context containing the API key.
func ContextWithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyContextKey, apiKey)
}

// APIKeyFromContext retrieves the API Key from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The API key if found.
//   - bool: True if the API key exists, false otherwise.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(APIKeyContextKey).(string)
	return val, ok
}

// ContextWithUser returns a new context with the user ID embedded.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - userID: string. The user ID to store.
//
// Returns:
//   - context.Context: A new context containing the user ID.
func ContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// UserFromContext retrieves the user ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The user ID if found.
//   - bool: True if the user ID exists, false otherwise.
func UserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserContextKey).(string)
	return val, ok
}

// ContextWithProfileID returns a new context with the profile ID embedded.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - profileID: string. The profile ID to store.
//
// Returns:
//   - context.Context: A new context containing the profile ID.
func ContextWithProfileID(ctx context.Context, profileID string) context.Context {
	return context.WithValue(ctx, ProfileIDContextKey, profileID)
}

// ProfileIDFromContext retrieves the profile ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The profile ID if found.
//   - bool: True if the profile ID exists, false otherwise.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ProfileIDContextKey).(string)
	return val, ok
}

// Authenticator defines the interface for authentication mechanisms.
type Authenticator interface {
	// Authenticate validates the request credentials.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - r: *http.Request. The HTTP request to authenticate.
	//
	// Returns:
	//   - context.Context: The authenticated user's context.
	//   - error: An error if authentication fails.
	Authenticate(ctx context.Context, r *http.Request) (context.Context, error)
}

// APIKeyAuthenticator implements API key authentication.
//
// It validates requests by checking for a configured API key in the specified location.
type APIKeyAuthenticator struct {
	ParamName string
	In        configv1.APIKeyAuth_Location
	Value     string
}

// NewAPIKeyAuthenticator initializes a new APIKeyAuthenticator.
//
// Parameters:
//   - config: *configv1.APIKeyAuth. The API key configuration.
//
// Returns:
//   - *APIKeyAuthenticator: The initialized authenticator.
//   - *APIKeyAuthenticator: nil if the configuration is invalid.
func NewAPIKeyAuthenticator(config *configv1.APIKeyAuth) *APIKeyAuthenticator {
	if config == nil || config.GetParamName() == "" || config.GetVerificationValue() == "" {
		return nil
	}
	return &APIKeyAuthenticator{
		ParamName: config.GetParamName(),
		In:        config.GetIn(),
		Value:     config.GetVerificationValue(),
	}
}

// Authenticate verifies the API key in the request.
//
// It checks if the parameter specified by `ParamName` matches the expected `Value`.
//
// Parameters:
//   - ctx: The request context.
//   - r: The HTTP request to authenticate.
//
// Returns:
//   - context.Context: The context with the API key (if valid).
//   - error: An error if the API key is missing or invalid.
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

	if subtle.ConstantTimeCompare([]byte(receivedKey), []byte(a.Value)) == 1 {
		return ContextWithAPIKey(ctx, receivedKey), nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// BasicAuthenticator authenticates using HTTP Basic Auth and bcrypt password hashing.
type BasicAuthenticator struct {
	PasswordHash string
	Username     string
}

// NewBasicAuthenticator initializes a new BasicAuthenticator.
//
// Parameters:
//   - config: *configv1.BasicAuth. The basic auth configuration.
//
// Returns:
//   - *BasicAuthenticator: The initialized authenticator.
//   - *BasicAuthenticator: nil if the configuration is invalid.
func NewBasicAuthenticator(config *configv1.BasicAuth) *BasicAuthenticator {
	if config == nil || config.GetPasswordHash() == "" {
		return nil
	}
	return &BasicAuthenticator{
		PasswordHash: config.GetPasswordHash(),
		Username:     config.GetUsername(),
	}
}

// Authenticate validates the basic auth credentials.
//
// Parameters:
//   - ctx: The request context.
//   - r: The HTTP request to authenticate.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: An error if credentials are missing or invalid.
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

	if usernameMatch && passwordMatch {
		return ctx, nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// TrustedHeaderAuthenticator authenticates using a trusted header (e.g., from an auth proxy).
type TrustedHeaderAuthenticator struct {
	HeaderName  string
	HeaderValue string // Optional: if empty, just checks presence
}

// NewTrustedHeaderAuthenticator initializes a new TrustedHeaderAuthenticator.
//
// Parameters:
//   - config: *configv1.TrustedHeaderAuth. The trusted header configuration.
//
// Returns:
//   - *TrustedHeaderAuthenticator: The initialized authenticator.
//   - *TrustedHeaderAuthenticator: nil if the configuration is invalid.
func NewTrustedHeaderAuthenticator(config *configv1.TrustedHeaderAuth) *TrustedHeaderAuthenticator {
	if config == nil || config.GetHeaderName() == "" {
		return nil
	}
	return &TrustedHeaderAuthenticator{
		HeaderName:  config.GetHeaderName(),
		HeaderValue: config.GetHeaderValue(),
	}
}

// Authenticate validates the trusted header.
//
// Parameters:
//   - ctx: The request context.
//   - r: The HTTP request to authenticate.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: An error if the header is missing or does not match the expected value.
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
	}
	return ctx, nil
}

// Manager manages authentication for the server.
//
// It maintains a registry of authenticators per service and delegates validation checks.
type Manager struct {
	authenticators *xsync.Map[string, Authenticator]
	apiKey         string

	// usersMu protects users map to allow atomic updates (hot-swap).
	usersMu sync.RWMutex
	users   map[string]*configv1.User

	// mu protects storage
	mu      sync.RWMutex
	storage storage.Storage
}

// NewManager initializes a new Authentication Manager.
//
// Parameters:
//   None.
//
// Returns:
//   - *Manager: The initialized manager instance.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
		users:          make(map[string]*configv1.User),
	}
}

// SetUsers updates the list of authorized users.
//
// Parameters:
//   - users: []*configv1.User. The list of user configurations.
func (am *Manager) SetUsers(users []*configv1.User) {
	am.usersMu.Lock()
	defer am.usersMu.Unlock()
	for _, u := range users {
		am.users[u.GetId()] = u
	}
}

// SetStorage configures the storage backend.
//
// Parameters:
//   - s: storage.Storage. The storage implementation.
func (am *Manager) SetStorage(s storage.Storage) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.storage = s
}

// GetUser retrieves a user configuration by ID.
//
// Parameters:
//   - id: string. The user identifier.
//
// Returns:
//   - *configv1.User: The user configuration.
//   - bool: True if found, false otherwise.
func (am *Manager) GetUser(id string) (*configv1.User, bool) {
	am.usersMu.RLock()
	defer am.usersMu.RUnlock()
	u, ok := am.users[id]
	return u, ok
}

// SetAPIKey configures the global API key.
//
// Parameters:
//   - apiKey: string. The API key to set.
func (am *Manager) SetAPIKey(apiKey string) {
	am.apiKey = apiKey
}

// AddAuthenticator registers an authenticator for a specific service.
//
// It overwrites any existing authenticator for the service.
//
// Parameters:
//   - serviceID: string. The service identifier.
//   - authenticator: Authenticator. The authenticator instance.
//
// Returns:
//   - error: An error if the authenticator is nil.
func (am *Manager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if authenticator == nil {
		return fmt.Errorf("authenticator for service %s is nil", serviceID)
	}
	am.authenticators.Store(serviceID, authenticator)
	return nil
}

// Authenticate validates request credentials for a service.
//
// It delegates to the service-specific authenticator if present, otherwise checks global credentials.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - serviceID: string. The service identifier.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: An error if authentication fails.
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
	}

	// Fallback: Check Global User Basic Auth
	if ctx, err := am.checkBasicAuthWithUsers(ctx, r); err == nil {
		return ctx, nil
	}

	// Otherwise, Fail Closed.
	return ctx, fmt.Errorf("unauthorized: no authentication configured")
}

// GetAuthenticator retrieves the authenticator for a service.
//
// Parameters:
//   - serviceID: string. The service identifier.
//
// Returns:
//   - Authenticator: The authenticator instance.
//   - bool: True if found, false otherwise.
func (am *Manager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	return am.authenticators.Load(serviceID)
}

// RemoveAuthenticator removes the authenticator for a service.
//
// Parameters:
//   - serviceID: string. The service identifier.
func (am *Manager) RemoveAuthenticator(serviceID string) {
	am.authenticators.Delete(serviceID)
}

// AddOAuth2Authenticator registers an OAuth2 authenticator for a service.
//
// This is a helper to initialize and register an OAuth2Authenticator.
//
// Parameters:
//   - ctx: context.Context. Context for OIDC provider initialization.
//   - serviceID: string. The service identifier.
//   - config: *OAuth2Config. The OAuth2 configuration.
//
// Returns:
//   - error: An error if initialization fails.
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

var (
	// oauthAuthenticatorCache stores *OAuth2Authenticator keyed by IssuerURL + joined Audiences.
	oauthAuthenticatorCache = xsync.NewMap[string, *OAuth2Authenticator]()
)

// ValidateAuthentication validates a request against a specific authentication configuration.
//
// It supports API Key, OAuth2, Basic Auth, and Trusted Header methods.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - config: *configv1.Authentication. The authentication configuration.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - error: An error if validation fails.
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

// checkBasicAuthWithUsers validates Basic Auth credentials against configured users.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: An error if authentication fails.
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
