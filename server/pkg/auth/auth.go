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

// ContextWithAPIKey returns a new context with the API Key.
func ContextWithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyContextKey, apiKey)
}

// APIKeyFromContext returns the API Key from the context.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(APIKeyContextKey).(string)
	return val, ok
}

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

// Authenticator checks if a request is authenticated.
type Authenticator interface {
	// Authenticate returns the authenticated user's context or an error.
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
	if config == nil || config.GetParamName() == "" || config.GetVerificationValue() == "" {
		return nil
	}
	return &APIKeyAuthenticator{
		ParamName: config.GetParamName(),
		In:        config.GetIn(),
		Value:     config.GetVerificationValue(),
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
// Authenticate verifies the API key in the request.
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

// NewBasicAuthenticator creates a new BasicAuthenticator.
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
func (a *BasicAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	user, password, ok := r.BasicAuth()
	if !ok {
		return ctx, fmt.Errorf("unauthorized")
	}

	if a.Username != "" && user != a.Username {
		return ctx, fmt.Errorf("unauthorized")
	}

	if passhash.CheckPassword(password, a.PasswordHash) {
		return ctx, nil
	}
	return ctx, fmt.Errorf("unauthorized")
}

// TrustedHeaderAuthenticator authenticates using a trusted header (e.g., from an auth proxy).
type TrustedHeaderAuthenticator struct {
	HeaderName  string
	HeaderValue string // Optional: if empty, just checks presence
}

// NewTrustedHeaderAuthenticator creates a new TrustedHeaderAuthenticator.
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
func (a *TrustedHeaderAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	val := r.Header.Get(a.HeaderName)
	if val == "" {
		return ctx, fmt.Errorf("unauthorized")
	}
	// If HeaderValue is set, it must match.
	if a.HeaderValue != "" {
		if val != a.HeaderValue {
			return ctx, fmt.Errorf("unauthorized")
		}
	}
	return ctx, nil
}

// Manager oversees the authentication process for the server. It maintains a
// registry of authenticators, each associated with a specific service ID, and
// delegates the authentication of requests to the appropriate authenticator.
// This allows for different authentication strategies to be used for different
// services.
type Manager struct {
	authenticators *xsync.Map[string, Authenticator]
	apiKey         string

	// usersMu protects users map to allow atomic updates (hot-swap).
	usersMu sync.RWMutex
	users   map[string]*configv1.User

	// mu protects storage
	// mu protects storage
	mu      sync.RWMutex
	storage storage.Storage
}

// NewManager creates and initializes a new Manager with an empty
// authenticator registry. This manager can then be used to register and manage
// authenticators for various services.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
		users:          make(map[string]*configv1.User),
	}
}

// SetUsers sets the users.
func (am *Manager) SetUsers(users []*configv1.User) {
	am.usersMu.Lock()
	defer am.usersMu.Unlock()
	for _, u := range users {
		am.users[u.GetId()] = u
	}
}

// SetStorage sets the storage.
func (am *Manager) SetStorage(s storage.Storage) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.storage = s
}

// GetUser retrieves a user by ID.
func (am *Manager) GetUser(id string) (*configv1.User, bool) {
	am.usersMu.RLock()
	defer am.usersMu.RUnlock()
	u, ok := am.users[id]
	return u, ok
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
			return ctx, fmt.Errorf("unauthorized")
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-API-Key")), []byte(am.apiKey)) != 1 {
			return ctx, fmt.Errorf("unauthorized")
		}
		ctx = ContextWithAPIKey(ctx, r.Header.Get("X-API-Key"))
	} else {
		// 🛡️ Sentinel Security Update: Fail Closed
		// If no global API key is set, and we are not using a service-specific authenticator below,
		// we must ensure we don't accidentally allow public access unless explicitly intended.
		// However, legitimate "public" services might exist if they have a "PublicAuthenticator".
		// But here we are handling the case where NO authenticator is found.
	}

	if authenticator, ok := am.authenticators.Load(serviceID); ok {
		return authenticator.Authenticate(ctx, r)
	}

	// 🛡️ Sentinel Security Update: Fail Closed
	// If a global API key was set and verified above, we allow the request (it fell through).
	if am.apiKey != "" {
		return ctx, nil
	}

	// If NO global key is set, and NO service authenticator is found, DENY.
	// This prevents "Fail Open" behavior where forgetting to configure auth leaves the service public.
	return ctx, fmt.Errorf("unauthorized: no authentication configured")
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

var (
	// oauthAuthenticatorCache stores *OAuth2Authenticator keyed by IssuerURL + joined Audiences.
	oauthAuthenticatorCache = xsync.NewMap[string, *OAuth2Authenticator]()
)

// ValidateAuthentication validates the authentication request against the provided configuration.
// It supports API Key and OAuth2 authentication methods.
//
// Parameters:
//   - ctx: The context for the request.
//   - config: The authentication configuration.
//   - r: The HTTP request to validate.
//
// Returns an error if validation fails or the method is unsupported.
func ValidateAuthentication(ctx context.Context, config *configv1.Authentication, r *http.Request) error {
	if config == nil {
		return nil // No auth configured implies allowed
	}

	switch method := config.AuthMethod.(type) {
	case *configv1.Authentication_ApiKey:
		authenticator := NewAPIKeyAuthenticator(method.ApiKey)
		if authenticator == nil {
			return fmt.Errorf("invalid API key configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case *configv1.Authentication_Oauth2:
		cfg := method.Oauth2
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
	case *configv1.Authentication_BasicAuth:
		authenticator := NewBasicAuthenticator(method.BasicAuth)
		if authenticator == nil {
			return fmt.Errorf("invalid Basic Auth configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case *configv1.Authentication_TrustedHeader:
		authenticator := NewTrustedHeaderAuthenticator(method.TrustedHeader)
		if authenticator == nil {
			return fmt.Errorf("invalid Trusted Header configuration")
		}
		_, err := authenticator.Authenticate(ctx, r)
		return err
	case *configv1.Authentication_Oidc:
		cfg := method.Oidc
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
		return nil
	}
}
