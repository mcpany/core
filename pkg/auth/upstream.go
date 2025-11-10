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

package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/valyala/fasttemplate"

	configv1 "github.com/mcpany/core/proto/config/v1"
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

// NewUpstreamAuthenticator creates an `UpstreamAuthenticator` based on the
// provided authentication configuration. It supports API key, bearer token, and
// basic authentication, as well as substitution of environment variables in the
// authentication parameters.
//
// If the `authConfig` is `nil`, no authenticator is created, and the function
// returns `nil, nil`. If the configuration is invalid (e.g., missing required
// fields), an error is returned.
//
// Parameters:
//   - authConfig: The configuration that specifies the authentication method
//     and its parameters.
//
// Returns an `UpstreamAuthenticator` or an error if the configuration is
// invalid.
func NewUpstreamAuthenticator(authConfig *configv1.UpstreamAuthentication) (UpstreamAuthenticator, error) {
	if authConfig == nil {
		return nil, nil
	}

	if authConfig.GetUseEnvironmentVariable() {
		err := substituteEnvVars(authConfig)
		if err != nil {
			return nil, err
		}
	}

	if apiKey := authConfig.GetApiKey(); apiKey != nil {
		apiKeyValue, err := ResolveSecretValue(apiKey.GetApiKey())
		if err != nil {
			return nil, err
		}
		if apiKey.GetHeaderName() == "" || apiKeyValue == "" {
			return nil, errors.New("API key authentication requires a header name and a key")
		}
		return &APIKeyAuth{
			HeaderName:  apiKey.GetHeaderName(),
			HeaderValue: apiKeyValue,
		}, nil
	}

	if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
		tokenValue, err := ResolveSecretValue(bearerToken.GetToken())
		if err != nil {
			return nil, err
		}
		if tokenValue == "" {
			return nil, errors.New("bearer token authentication requires a token")
		}
		return &BearerTokenAuth{
			Token: tokenValue,
		}, nil
	}

	if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
		username, err := ResolveSecretValue(basicAuth.GetUsername())
		if err != nil {
			return nil, err
		}
		password, err := ResolveSecretValue(basicAuth.GetPassword())
		if err != nil {
			return nil, err
		}
		if username == "" {
			return nil, errors.New("basic authentication requires a username")
		}
		return &BasicAuth{
			Username: username,
			Password: password,
		}, nil
	}

	return nil, nil
}

func substituteEnvVars(authConfig *configv1.UpstreamAuthentication) error {
	envVars := make(map[string]interface{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envVars[pair[0]] = pair[1]
	}

	applySubstitution := func(secret *configv1.SecretValue) (*configv1.SecretValue, error) {
		if secret == nil {
			return nil, nil
		}
		val, err := ResolveSecretValue(secret)
		if err != nil {
			return nil, err
		}
		substitutedVal := fasttemplate.New(val, "{{", "}}").ExecuteString(envVars)
		return configv1.SecretValue_builder{PlainText: &substitutedVal}.Build(), nil
	}

	if apiKey := authConfig.GetApiKey(); apiKey != nil {
		apiKey.SetHeaderName(fasttemplate.New(apiKey.GetHeaderName(), "{{", "}}").ExecuteString(envVars))
		newSecret, err := applySubstitution(apiKey.GetApiKey())
		if err != nil {
			return err
		}
		apiKey.SetApiKey(newSecret)
	}
	if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
		newSecret, err := applySubstitution(bearerToken.GetToken())
		if err != nil {
			return err
		}
		bearerToken.SetToken(newSecret)
	}
	if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
		newUsername, err := applySubstitution(basicAuth.GetUsername())
		if err != nil {
			return err
		}
		basicAuth.SetUsername(newUsername)

		newPassword, err := applySubstitution(basicAuth.GetPassword())
		if err != nil {
			return err
		}
		basicAuth.SetPassword(newPassword)
	}
	return nil
}

// APIKeyAuth implements UpstreamAuthenticator for API key-based authentication.
// It adds a specified header with a static API key value to the request.
type APIKeyAuth struct {
	HeaderName  string
	HeaderValue string
}

// Authenticate adds the configured API key to the request's header.
//
// Parameters:
//   - req: The HTTP request to be modified.
//
// Returns `nil` on success.
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
// Parameters:
//   - req: The HTTP request to be modified.
//
// Returns `nil` on success.
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
// Parameters:
//   - req: The HTTP request to be modified.
//
// Returns `nil` on success.
func (b *BasicAuth) Authenticate(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}
