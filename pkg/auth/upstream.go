/*
 * Copyright 2025 Author(s) of MCPXY
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
	"errors"
	"net/http"

	configv1 "github.com/mcpxy/core/proto/config/v1"
)

// UpstreamAuthenticator defines the interface for authentication methods used
// when communicating with upstream services. Each implementation is responsible
// for modifying the HTTP request to include the necessary authentication
// credentials.
type UpstreamAuthenticator interface {
	// Authenticate modifies the given HTTP request to add authentication
	// information, such as headers or basic auth credentials.
	Authenticate(req *http.Request) error
}

// NewUpstreamAuthenticator creates an UpstreamAuthenticator based on the
// provided authentication configuration. It supports API key, bearer token, and
// basic authentication.
//
// authConfig is the configuration that specifies the authentication method and
// its parameters. If authConfig is nil, no authenticator is created.
// It returns an UpstreamAuthenticator or an error if the configuration is
// invalid.
func NewUpstreamAuthenticator(authConfig *configv1.UpstreamAuthentication) (UpstreamAuthenticator, error) {
	if authConfig == nil {
		return nil, nil
	}

	if apiKey := authConfig.GetApiKey(); apiKey != nil {
		if apiKey.GetHeaderName() == "" || apiKey.GetApiKey() == "" {
			return nil, errors.New("API key authentication requires a header name and a key")
		}
		return &APIKeyAuth{
			HeaderName:  apiKey.GetHeaderName(),
			HeaderValue: apiKey.GetApiKey(),
		}, nil
	}

	if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
		if bearerToken.GetToken() == "" {
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
		return &BasicAuth{
			Username: basicAuth.GetUsername(),
			Password: basicAuth.GetPassword(),
		}, nil
	}

	return nil, nil
}

// APIKeyAuth implements UpstreamAuthenticator for API key-based authentication.
// It adds a specified header with a static API key value to the request.
type APIKeyAuth struct {
	HeaderName  string
	HeaderValue string
}

// Authenticate adds the configured API key to the request's header.
//
// req is the HTTP request to be modified.
func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	req.Header.Set(a.HeaderName, a.HeaderValue)
	return nil
}

// BearerTokenAuth implements UpstreamAuthenticator for bearer token-based
// authentication. It adds an "Authorization" header with a bearer token.
type BearerTokenAuth struct {
	Token string
}

// Authenticate adds the bearer token to the request's "Authorization" header.
//
// req is the HTTP request to be modified.
func (b *BearerTokenAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.Token)
	return nil
}

// BasicAuth implements UpstreamAuthenticator for basic HTTP authentication.
// It adds an "Authorization" header with the username and password.
type BasicAuth struct {
	Username string
	Password string
}

// Authenticate sets the request's basic authentication credentials.
//
// req is the HTTP request to be modified.
func (b *BasicAuth) Authenticate(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}
