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
// Summary: Embeds an API key into the context.
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

// APIKeyFromContext returns the API Key from the context if present.
//
// Summary: Retrieves the API key from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The API key.
//   - bool: True if found.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(APIKeyContextKey).(string)
	return val, ok
}

// ContextWithUser returns a new context with the user ID embedded.
//
// Summary: Embeds a user ID into the context.
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

// UserFromContext returns the user ID from the context if present.
//
// Summary: Retrieves the user ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The user ID.
//   - bool: True if found.
func UserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserContextKey).(string)
	return val, ok
}

// ContextWithProfileID returns a new context with the profile ID embedded.
//
// Summary: Embeds a profile ID into the context.
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

// ProfileIDFromContext returns the profile ID from the context if present.
//
// Summary: Retrieves the profile ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - string: The profile ID.
//   - bool: True if found.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ProfileIDContextKey).(string)
	return val, ok
}

// Authenticator defines the interface for authentication mechanisms.
//
// Summary: Interface for authenticating HTTP requests.
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
	//   - context.Context: The authenticated context (e.g. with user info).
	//   - error: An error if authentication fails.
	Authenticate(ctx context.Context, r *http.Request) (context.Context, error)
}

// APIKeyAuthenticator provides an authentication mechanism based on a static API key.
//
// Summary: Authenticates requests using a static API key.
type APIKeyAuthenticator struct {
	ParamName string
	In        configv1.APIKeyAuth_Location
	Value     string
}

// NewAPIKeyAuthenticator creates a new APIKeyAuthenticator instance.
//
// Summary: Initializes an APIKeyAuthenticator.
//
// Parameters:
//   - config: *configv1.APIKeyAuth. The configuration settings.
//
// Returns:
//   - *APIKeyAuthenticator: The initialized authenticator, or nil if config is invalid.
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
// Summary: Validates the API key from header, query, or cookie.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Context with API key if valid.
//   - error: Error if unauthorized.
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
//
// Summary: Authenticates requests using HTTP Basic Auth.
type BasicAuthenticator struct {
	PasswordHash string
	Username     string
}

// NewBasicAuthenticator creates a new BasicAuthenticator instance.
//
// Summary: Initializes a BasicAuthenticator.
//
// Parameters:
//   - config: *configv1.BasicAuth. The configuration settings.
//
// Returns:
//   - *BasicAuthenticator: The initialized authenticator, or nil if config is invalid.
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
// Summary: Validates username and password hash.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Authenticated context.
//   - error: Error if unauthorized.
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

// TrustedHeaderAuthenticator authenticates using a trusted header.
//
// Summary: Authenticates requests based on the presence/value of a specific header.
type TrustedHeaderAuthenticator struct {
	HeaderName  string
	HeaderValue string // Optional: if empty, just checks presence
}

// NewTrustedHeaderAuthenticator creates a new TrustedHeaderAuthenticator instance.
//
// Summary: Initializes a TrustedHeaderAuthenticator.
//
// Parameters:
//   - config: *configv1.TrustedHeaderAuth. The configuration settings.
//
// Returns:
//   - *TrustedHeaderAuthenticator: The initialized authenticator, or nil if config is invalid.
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
// Summary: Checks for the trusted header.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: Authenticated context.
//   - error: Error if unauthorized.
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

// Manager oversees the authentication process for the server.
//
// Summary: Manages authentication strategies and user sessions.
type Manager struct {
	authenticators *xsync.Map[string, Authenticator]
	apiKey         string

	// dummyHash is used to prevent timing attacks during user enumeration.
	dummyHash string

	// usersMu protects users map to allow atomic updates (hot-swap).
	usersMu sync.RWMutex
	users   map[string]*configv1.User

	// mu protects storage
	mu      sync.RWMutex
	storage storage.Storage
}

var (
	// dummyHash is used to prevent timing attacks during user enumeration.
	// It is computed lazily once.
	dummyHash     string
	dummyHashOnce sync.Once
)

// getDummyHash returns the cached dummy hash.
func getDummyHash() string {
	dummyHashOnce.Do(func() {
		// Pre-calculate a dummy hash for timing mitigation.
		// We use a fixed string. The cost factor is determined by passhash.Password default (12).
		var err error
		dummyHash, err = passhash.Password("dummy-password-for-timing-mitigation")
		if err != nil {
			// This should practically never happen unless system is OOM.
			panic(fmt.Sprintf("failed to generate dummy hash: %v", err))
		}
	})
	return dummyHash
}

// NewManager creates and initializes a new Manager with an empty authenticator registry.
//
// Summary: Initializes a new Authentication Manager.
//
// Returns:
//   - *Manager: A new Manager instance.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
		users:          make(map[string]*configv1.User),
		dummyHash:      getDummyHash(),
	}
}

// SetUsers updates the list of active users.
//
// Summary: Sets the configured users.
//
// Parameters:
//   - users: []*configv1.User. The list of users.
func (am *Manager) SetUsers(users []*configv1.User) {
	am.usersMu.Lock()
	defer am.usersMu.Unlock()
	for _, u := range users {
		am.users[u.GetId()] = u
	}
}

// SetStorage sets the storage backend for the manager.
//
// Summary: Configures the storage backend.
//
// Parameters:
//   - s: storage.Storage. The storage implementation.
func (am *Manager) SetStorage(s storage.Storage) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.storage = s
}

// GetUser retrieves a user configuration by their ID.
//
// Summary: Looks up a user by ID.
//
// Parameters:
//   - id: string. The user ID.
//
// Returns:
//   - *configv1.User: The user configuration.
//   - bool: True if found.
func (am *Manager) GetUser(id string) (*configv1.User, bool) {
	am.usersMu.RLock()
	defer am.usersMu.RUnlock()
	u, ok := am.users[id]
	return u, ok
}

// SetAPIKey sets the global API key for the server.
//
// Summary: Sets the global API key.
//
// Parameters:
//   - apiKey: string. The API key.
func (am *Manager) SetAPIKey(apiKey string) {
	am.apiKey = apiKey
}

// AddAuthenticator registers an authenticator for a given service ID.
//
// Summary: Registers an authenticator for a service.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - authenticator: Authenticator. The authenticator instance.
//
// Returns:
//   - error: Error if authenticator is nil.
func (am *Manager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if authenticator == nil {
		return fmt.Errorf("authenticator for service %s is nil", serviceID)
	}
	am.authenticators.Store(serviceID, authenticator)
	return nil
}

// Authenticate authenticates a request for a specific service.
//
// Summary: Authenticates a request, checking service-specific or global rules.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - serviceID: string. The service ID.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - context.Context: The authenticated context.
//   - error: Error if unauthorized.
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

// GetAuthenticator retrieves the authenticator registered for a specific service.
//
// Summary: Looks up an authenticator by service ID.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
//   - Authenticator: The authenticator instance.
//   - bool: True if found.
func (am *Manager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	return am.authenticators.Load(serviceID)
}

// RemoveAuthenticator removes the authenticator for a given service ID.
//
// Summary: Removes an authenticator by service ID.
//
// Parameters:
//   - serviceID: string. The service ID.
func (am *Manager) RemoveAuthenticator(serviceID string) {
	am.authenticators.Delete(serviceID)
}

// AddOAuth2Authenticator creates and registers a new OAuth2Authenticator for a given service ID.
//
// Summary: Helper to add an OAuth2 authenticator.
//
// Parameters:
//   - ctx: context.Context. Context for initialization.
//   - serviceID: string. The service ID.
//   - config: *OAuth2Config. The OAuth2 configuration.
//
// Returns:
//   - error: Error if creation fails.
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

	var targetUser *configv1.User
	// Initialize with dummy hash to ensure constant time execution even if user is not found.
	targetHash := am.dummyHash
	// validUserFound tracks if we found a user AND they have Basic Auth configured.
	// This prevents logging in as a user who exists but doesn't have Basic Auth (using the dummy password).
	validUserFound := false

	// Direct lookup if user ID matches username
	if user, ok := am.users[username]; ok {
		targetUser = user
		if basicAuth := user.GetAuthentication().GetBasicAuth(); basicAuth != nil {
			targetHash = basicAuth.GetPasswordHash()
			validUserFound = true
		}
	}

	// Always check password to avoid timing attacks (User Enumeration).
	// If user is not found, we check against dummyHash.
	// If user is found, we check against their hash.
	// passhash.CheckPassword (bcrypt) takes ~300ms (cost 12), masking the map lookup time.
	passwordMatch := passhash.CheckPassword(password, targetHash)

	if validUserFound && passwordMatch {
		ctx = ContextWithUser(ctx, targetUser.GetId())
		if len(targetUser.GetRoles()) > 0 {
			ctx = ContextWithRoles(ctx, targetUser.GetRoles())
		}
		return ctx, nil
	}

	// Fallback: Iterate all users (in case username is not ID, but we assume ID==Username for now)
	return ctx, fmt.Errorf("invalid credentials")
}
