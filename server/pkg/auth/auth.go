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
// Summary: UserContextKey is the context key for the user ID. Defines UserContextKey.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	UserContextKey authContextKey = "user_id"
// Summary: ProfileIDContextKey is the context key for the profile ID. Defines ProfileIDContextKey.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	ProfileIDContextKey authContextKey = "profile_id"
// Summary: APIKeyContextKey is the context key for the API Key. Defines APIKeyContextKey.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	APIKeyContextKey authContextKey = "api_key"
)

// Summary: ContextWithAPIKey returns a new context with the API Key embedded. Embeds an API key into the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - apiKey (string): The apiKey parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ContextWithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyContextKey, apiKey)
}

// Summary: APIKeyFromContext returns the API Key from the context if present. Retrieves the API key from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(APIKeyContextKey).(string)
	return val, ok
}

// Summary: ContextWithUser returns a new context with the user ID embedded. Embeds a user ID into the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - userID (string): The userID parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// Summary: UserFromContext returns the user ID from the context if present. Retrieves the user ID from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func UserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserContextKey).(string)
	return val, ok
}

// Summary: ContextWithProfileID returns a new context with the profile ID embedded. Embeds a profile ID into the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - profileID (string): The profileID parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ContextWithProfileID(ctx context.Context, profileID string) context.Context {
	return context.WithValue(ctx, ProfileIDContextKey, profileID)
}

// Summary: ProfileIDFromContext returns the profile ID from the context if present. Retrieves the profile ID from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ProfileIDContextKey).(string)
	return val, ok
}

// Summary: Authenticator defines the interface for authentication mechanisms. Interface for authenticating HTTP requests.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
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
	//   - context.Context: The authenticated context (e.g. with user info).
	//   - error: An error if authentication fails.
	Authenticate(ctx context.Context, r *http.Request) (context.Context, error)
}

// Summary: APIKeyAuthenticator provides an authentication mechanism based on a static API key. Authenticates requests using a static API key.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type APIKeyAuthenticator struct {
	ParamName string
	In        configv1.APIKeyAuth_Location
	Value     string
}

// Summary: NewAPIKeyAuthenticator creates a new APIKeyAuthenticator instance. Initializes an APIKeyAuthenticator.
//
// Parameters:
//   - config (*configv1.APIKeyAuth): The config parameter.
//
// Returns:
//   - *APIKeyAuthenticator: The resulting *APIKeyAuthenticator.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: Authenticate verifies the API key in the request. Validates the API key from header, query, or cookie.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - r (*http.Request): The r parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: BasicAuthenticator authenticates using HTTP Basic Auth and bcrypt password hashing. Authenticates requests using HTTP Basic Auth.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type BasicAuthenticator struct {
	PasswordHash string
	Username     string
}

// Summary: NewBasicAuthenticator creates a new BasicAuthenticator instance. Initializes a BasicAuthenticator.
//
// Parameters:
//   - config (*configv1.BasicAuth): The config parameter.
//
// Returns:
//   - *BasicAuthenticator: The resulting *BasicAuthenticator.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewBasicAuthenticator(config *configv1.BasicAuth) *BasicAuthenticator {
	if config == nil || config.GetPasswordHash() == "" {
		return nil
	}
	return &BasicAuthenticator{
		PasswordHash: config.GetPasswordHash(),
		Username:     config.GetUsername(),
	}
}

// Summary: Authenticate validates the basic auth credentials. Validates username and password hash.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - r (*http.Request): The r parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: TrustedHeaderAuthenticator authenticates using a trusted header. Authenticates requests based on the presence/value of a specific header.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type TrustedHeaderAuthenticator struct {
	HeaderName  string
	HeaderValue string // Optional: if empty, just checks presence
}

// Summary: NewTrustedHeaderAuthenticator creates a new TrustedHeaderAuthenticator instance. Initializes a TrustedHeaderAuthenticator.
//
// Parameters:
//   - config (*configv1.TrustedHeaderAuth): The config parameter.
//
// Returns:
//   - *TrustedHeaderAuthenticator: The resulting *TrustedHeaderAuthenticator.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewTrustedHeaderAuthenticator(config *configv1.TrustedHeaderAuth) *TrustedHeaderAuthenticator {
	if config == nil || config.GetHeaderName() == "" {
		return nil
	}
	return &TrustedHeaderAuthenticator{
		HeaderName:  config.GetHeaderName(),
		HeaderValue: config.GetHeaderValue(),
	}
}

// Summary: Authenticate validates the trusted header. Checks for the trusted header.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - r (*http.Request): The r parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: Manager oversees the authentication process for the server. Manages authentication strategies and user sessions.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: NewManager creates and initializes a new Manager with an empty authenticator registry. Initializes a new Authentication Manager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Manager: The resulting *Manager.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
		users:          make(map[string]*configv1.User),
	}
}

// Summary: SetUsers updates the list of active users. Sets the configured users.
//
// Parameters:
//   - users ([]*configv1.User): The users parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) SetUsers(users []*configv1.User) {
	am.usersMu.Lock()
	defer am.usersMu.Unlock()
	for _, u := range users {
		am.users[u.GetId()] = u
	}
}

// Summary: SetStorage sets the storage backend for the manager. Configures the storage backend.
//
// Parameters:
//   - s (storage.Storage): The s parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) SetStorage(s storage.Storage) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.storage = s
}

// Summary: GetUser retrieves a user configuration by their ID. Looks up a user by ID.
//
// Parameters:
//   - id (string): The id parameter.
//
// Returns:
//   - *configv1.User: The resulting *configv1.User.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) GetUser(id string) (*configv1.User, bool) {
	am.usersMu.RLock()
	defer am.usersMu.RUnlock()
	u, ok := am.users[id]
	return u, ok
}

// Summary: SetAPIKey sets the global API key for the server. Sets the global API key.
//
// Parameters:
//   - apiKey (string): The apiKey parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) SetAPIKey(apiKey string) {
	am.apiKey = apiKey
}

// Summary: AddAuthenticator registers an authenticator for a given service ID. Registers an authenticator for a service.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//   - authenticator (Authenticator): The authenticator parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (am *Manager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if authenticator == nil {
		return fmt.Errorf("authenticator for service %s is nil", serviceID)
	}
	am.authenticators.Store(serviceID, authenticator)
	return nil
}

// Summary: Authenticate authenticates a request for a specific service. Authenticates a request, checking service-specific or global rules.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - serviceID (string): The serviceID parameter.
//   - r (*http.Request): The r parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetAuthenticator retrieves the authenticator registered for a specific service. Looks up an authenticator by service ID.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - Authenticator: The resulting Authenticator.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	return am.authenticators.Load(serviceID)
}

// Summary: RemoveAuthenticator removes the authenticator for a given service ID. Removes an authenticator by service ID.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (am *Manager) RemoveAuthenticator(serviceID string) {
	am.authenticators.Delete(serviceID)
}

// Summary: AddOAuth2Authenticator creates and registers a new OAuth2Authenticator for a given service ID. Helper to add an OAuth2 authenticator.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - serviceID (string): The serviceID parameter.
//   - config (*OAuth2Config): The config parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: ValidateAuthentication validates the authentication request against the provided configuration. Validates a request against a specific auth configuration.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - config (*configv1.Authentication): The config parameter.
//   - r (*http.Request): The r parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
