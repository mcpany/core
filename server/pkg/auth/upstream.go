// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/mcpany/core/server/pkg/util"
	"golang.org/x/oauth2/clientcredentials"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// UpstreamAuthenticator defines the interface for authentication methods used.
//
// Summary: defines the interface for authentication methods used.
type UpstreamAuthenticator interface {
	// Authenticate modifies the given HTTP request to add authentication.
	//
	// Summary: modifies the given HTTP request to add authentication.
	//
	// Parameters:
	//   - req: *http.Request. The request object.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Authenticate(req *http.Request) error
}

// NewUpstreamAuthenticator creates an `UpstreamAuthenticator` based on the.
//
// Summary: creates an `UpstreamAuthenticator` based on the.
//
// Parameters:
//   - authConfig: *configv1.Authentication. The authConfig.
//
// Returns:
//   - UpstreamAuthenticator: The UpstreamAuthenticator.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewUpstreamAuthenticator(authConfig *configv1.Authentication) (UpstreamAuthenticator, error) {
	if authConfig == nil {
		return nil, nil
	}

	if apiKey := authConfig.GetApiKey(); apiKey != nil {
		if apiKey.GetParamName() == "" {
			return nil, errors.New("API key authentication requires a parameter name")
		}
		if apiKey.GetValue() == nil {
			return nil, errors.New("API key authentication requires an API key value")
		}
		return &APIKeyAuth{
			ParamName: apiKey.GetParamName(),
			Value:     apiKey.GetValue(),
			Location:  apiKey.GetIn(),
		}, nil
	}

	if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
		if bearerToken.GetToken() == nil {
			return nil, errors.New("bearer token authentication requires a token")
		}
		return &BearerTokenAuth{
			Token: bearerToken.GetToken(),
		}, nil
	}

	if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
		if basicAuth.GetUsername() == "" {
			return nil, errors.New("basic authentication requires a username")
		}
		if basicAuth.GetPassword() == nil {
			return nil, errors.New("basic authentication requires a password")
		}
		return &BasicAuth{
			Username: basicAuth.GetUsername(),
			Password: basicAuth.GetPassword(),
		}, nil
	}

	if oauth2 := authConfig.GetOauth2(); oauth2 != nil {
		if oauth2.GetClientId() == nil {
			return nil, errors.New("OAuth2 authentication requires a client ID")
		}
		if oauth2.GetClientSecret() == nil {
			return nil, errors.New("OAuth2 authentication requires a client secret")
		}
		if oauth2.GetTokenUrl() == "" && oauth2.GetIssuerUrl() == "" {
			return nil, errors.New("OAuth2 authentication requires a token URL or an issuer URL")
		}
		return &OAuth2Auth{
			ClientID:     oauth2.GetClientId(),
			ClientSecret: oauth2.GetClientSecret(),
			TokenURL:     oauth2.GetTokenUrl(),
			IssuerURL:    oauth2.GetIssuerUrl(),
			Scopes:       strings.Split(oauth2.GetScopes(), " "),
		}, nil
	}

	return nil, nil
}

// APIKeyAuth implements UpstreamAuthenticator for API key-based authentication.
//
// Summary: implements UpstreamAuthenticator for API key-based authentication.
type APIKeyAuth struct {
	ParamName string
	Value     *configv1.SecretValue
	Location  configv1.APIKeyAuth_Location
}

// Authenticate adds the configured API key to the request's header, query, or cookie.
//
// Summary: adds the configured API key to the request's header, query, or cookie.
//
// Parameters:
//   - req: *http.Request. The req.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	if a.Value == nil {
		return errors.New("api key secret is not configured")
	}
	value, err := util.ResolveSecret(req.Context(), a.Value)
	if err != nil {
		return err
	}

	switch a.Location {
	case configv1.APIKeyAuth_QUERY:
		q := req.URL.Query()
		q.Set(a.ParamName, value)
		req.URL.RawQuery = q.Encode()
	case configv1.APIKeyAuth_COOKIE:
		req.AddCookie(&http.Cookie{
			Name:  a.ParamName,
			Value: value,
		})
	case configv1.APIKeyAuth_HEADER:
		fallthrough
	default:
		req.Header.Set(a.ParamName, value)
	}
	return nil
}

// BearerTokenAuth implements UpstreamAuthenticator for bearer token-based.
//
// Summary: implements UpstreamAuthenticator for bearer token-based.
type BearerTokenAuth struct {
	Token *configv1.SecretValue
}

// Authenticate adds the bearer token to the request's "Authorization" header.
//
// Summary: adds the bearer token to the request's "Authorization" header.
//
// Parameters:
//   - req: *http.Request. The req.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (b *BearerTokenAuth) Authenticate(req *http.Request) error {
	if b.Token == nil {
		return errors.New("bearer token secret is not configured")
	}
	token, err := util.ResolveSecret(req.Context(), b.Token)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// BasicAuth implements UpstreamAuthenticator for basic HTTP authentication.
//
// Summary: implements UpstreamAuthenticator for basic HTTP authentication.
type BasicAuth struct {
	Username string
	Password *configv1.SecretValue
}

// Authenticate sets the request's basic authentication credentials.
//
// Summary: sets the request's basic authentication credentials.
//
// Parameters:
//   - req: *http.Request. The req.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (b *BasicAuth) Authenticate(req *http.Request) error {
	if b.Password == nil {
		return errors.New("basic auth password secret is not configured")
	}
	password, err := util.ResolveSecret(req.Context(), b.Password)
	if err != nil {
		return err
	}
	req.SetBasicAuth(b.Username, password)
	return nil
}

// OAuth2Auth implements UpstreamAuthenticator for OAuth2 client credentials flow.
//
// Summary: implements UpstreamAuthenticator for OAuth2 client credentials flow.
type OAuth2Auth struct {
	ClientID     *configv1.SecretValue
	ClientSecret *configv1.SecretValue
	TokenURL     string
	IssuerURL    string
	Scopes       []string

	discoveryMu sync.Mutex
}

// getTokenURL returns the token URL, performing discovery if necessary.
func (o *OAuth2Auth) getTokenURL(ctx context.Context) (string, error) {
	o.discoveryMu.Lock()
	defer o.discoveryMu.Unlock()

	if o.TokenURL != "" {
		return o.TokenURL, nil
	}

	if o.IssuerURL != "" {
		// Create a context for discovery
		provider, err := oidc.NewProvider(ctx, o.IssuerURL)
		if err != nil {
			return "", fmt.Errorf("failed to discover OIDC configuration from issuer %q: %w", o.IssuerURL, err)
		}
		o.TokenURL = provider.Endpoint().TokenURL
		return o.TokenURL, nil
	}

	return "", errors.New("OAuth2 authentication requires a token URL (and no issuer provided)")
}

// Authenticate fetches a token and adds it to the request's "Authorization" header.
//
// Summary: fetches a token and adds it to the request's "Authorization" header.
//
// Parameters:
//   - req: *http.Request. The req.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (o *OAuth2Auth) Authenticate(req *http.Request) error {
	tokenURL, err := o.getTokenURL(req.Context())
	if err != nil {
		return err
	}

	if o.ClientID == nil {
		return errors.New("oauth2 client id secret is not configured")
	}
	if o.ClientSecret == nil {
		return errors.New("oauth2 client secret is not configured")
	}
	clientID, err := util.ResolveSecret(req.Context(), o.ClientID)
	if err != nil {
		return err
	}
	clientSecret, err := util.ResolveSecret(req.Context(), o.ClientSecret)
	if err != nil {
		return err
	}
	cfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       o.Scopes,
	}
	token, err := cfg.TokenSource(req.Context()).Token()
	if err != nil {
		return err
	}
	token.SetAuthHeader(req)
	return nil
}
