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
	//
	// Summary: Defines UserContextKey.
	UserContextKey authContextKey = "user_id"
	// ProfileIDContextKey is the context key for the profile ID.
	//
	// Summary: Defines ProfileIDContextKey.
	ProfileIDContextKey authContextKey = "profile_id"
	// APIKeyContextKey is the context key for the API Key.
	//
	// Summary: Defines APIKeyContextKey.
	APIKeyContextKey authContextKey = "api_key"
)

// ContextWithAPIKey serves as a public interface for interacting with ContextWithAPIKey.
//
// Summary: Context the with api key appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func ContextWithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyContextKey, apiKey)
}

// APIKeyFromContext serves as a public interface for interacting with APIKeyFromContext.
//
// Summary: Api the key from context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(APIKeyContextKey).(string)
	return val, ok
}

// ContextWithUser serves as a public interface for interacting with ContextWithUser.
//
// Summary: Context the with user appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func ContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// UserFromContext serves as a public interface for interacting with UserFromContext.
//
// Summary: User the from context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func UserFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserContextKey).(string)
	return val, ok
}

// ContextWithProfileID serves as a public interface for interacting with ContextWithProfileID.
//
// Summary: Context the with profile id appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func ContextWithProfileID(ctx context.Context, profileID string) context.Context {
	return context.WithValue(ctx, ProfileIDContextKey, profileID)
}

// ProfileIDFromContext serves as a public interface for interacting with ProfileIDFromContext.
//
// Summary: Profile the id from context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func ProfileIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ProfileIDContextKey).(string)
	return val, ok
}

// Authenticator represents the public Authenticator entity.
//
// Summary: Defines the structured data model representing a .
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

// APIKeyAuthenticator represents the public APIKeyAuthenticator entity.
//
// Summary: Defines the structured data model representing a key authenticator.
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

// NewAPIKeyAuthenticator serves as a public interface for interacting with NewAPIKeyAuthenticator.
//
// Summary: Constructs and returns an initialized api key authenticator ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Authenticate serves as a public interface for interacting with Authenticate.
//
// Summary: Authenticate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// BasicAuthenticator represents the public BasicAuthenticator entity.
//
// Summary: Defines the structured data model representing a authenticator.
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

// NewBasicAuthenticator serves as a public interface for interacting with NewBasicAuthenticator.
//
// Summary: Constructs and returns an initialized basic authenticator ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewBasicAuthenticator(config *configv1.BasicAuth) *BasicAuthenticator {
	if config == nil || config.GetPasswordHash() == "" {
		return nil
	}
	return &BasicAuthenticator{
		PasswordHash: config.GetPasswordHash(),
		Username:     config.GetUsername(),
	}
}

// Authenticate serves as a public interface for interacting with Authenticate.
//
// Summary: Authenticate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// TrustedHeaderAuthenticator represents the public TrustedHeaderAuthenticator entity.
//
// Summary: Defines the structured data model representing a header authenticator.
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

// NewTrustedHeaderAuthenticator serves as a public interface for interacting with NewTrustedHeaderAuthenticator.
//
// Summary: Constructs and returns an initialized trusted header authenticator ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewTrustedHeaderAuthenticator(config *configv1.TrustedHeaderAuth) *TrustedHeaderAuthenticator {
	if config == nil || config.GetHeaderName() == "" {
		return nil
	}
	return &TrustedHeaderAuthenticator{
		HeaderName:  config.GetHeaderName(),
		HeaderValue: config.GetHeaderValue(),
	}
}

// Authenticate serves as a public interface for interacting with Authenticate.
//
// Summary: Authenticate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
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

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewManager() *Manager {
	return &Manager{
		authenticators: xsync.NewMap[string, Authenticator](),
		users:          make(map[string]*configv1.User),
	}
}

// SetUsers serves as a public interface for interacting with SetUsers.
//
// Summary: Set the users appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) SetUsers(users []*configv1.User) {
	am.usersMu.Lock()
	defer am.usersMu.Unlock()
	for _, u := range users {
		am.users[u.GetId()] = u
	}
}

// SetStorage serves as a public interface for interacting with SetStorage.
//
// Summary: Set the storage appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) SetStorage(s storage.Storage) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.storage = s
}

// GetUser serves as a public interface for interacting with GetUser.
//
// Summary: Fetches and returns the underlying user from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) GetUser(id string) (*configv1.User, bool) {
	am.usersMu.RLock()
	defer am.usersMu.RUnlock()
	u, ok := am.users[id]
	return u, ok
}

// SetAPIKey serves as a public interface for interacting with SetAPIKey.
//
// Summary: Set the api key appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) SetAPIKey(apiKey string) {
	am.apiKey = apiKey
}

// AddAuthenticator serves as a public interface for interacting with AddAuthenticator.
//
// Summary: Add the authenticator appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if authenticator == nil {
		return fmt.Errorf("authenticator for service %s is nil", serviceID)
	}
	am.authenticators.Store(serviceID, authenticator)
	return nil
}

// Authenticate serves as a public interface for interacting with Authenticate.
//
// Summary: Authenticate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetAuthenticator serves as a public interface for interacting with GetAuthenticator.
//
// Summary: Fetches and returns the underlying authenticator from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	return am.authenticators.Load(serviceID)
}

// RemoveAuthenticator serves as a public interface for interacting with RemoveAuthenticator.
//
// Summary: Remove the authenticator appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (am *Manager) RemoveAuthenticator(serviceID string) {
	am.authenticators.Delete(serviceID)
}

// AddOAuth2Authenticator serves as a public interface for interacting with AddOAuth2Authenticator.
//
// Summary: Add the o auth authenticator appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ValidateAuthentication serves as a public interface for interacting with ValidateAuthentication.
//
// Summary: Validate the authentication appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
