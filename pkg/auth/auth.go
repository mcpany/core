/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law-or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mcpany/core/pkg/consts"
)

// Authenticator is the interface for all authentication methods.
type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (context.Context, error)
}

// APIKeyAuthenticator authenticates requests using an API key.
type APIKeyAuthenticator struct {
	HeaderName  string
	HeaderValue string
}

// NewAPIKeyAuthenticator creates a new APIKeyAuthenticator.
func NewAPIKeyAuthenticator(paramName, keyValue string) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{
		HeaderName:  paramName,
		HeaderValue: keyValue,
	}
}

// Authenticate authenticates the request by checking the API key in the header.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, req *http.Request) (context.Context, error) {
	providedKey := req.Header.Get(a.HeaderName)
	if providedKey == "" {
		return nil, fmt.Errorf("API key is missing")
	}

	if providedKey != a.HeaderValue {
		return nil, fmt.Errorf("invalid API key")
	}

	return context.WithValue(ctx, consts.HeaderAPIKey, providedKey), nil
}

// AuthManager manages the authenticators for different services.
type AuthManager struct {
	authenticators map[string]Authenticator
	apiKey         string
}

// NewAuthManager creates a new AuthManager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		authenticators: make(map[string]Authenticator),
	}
}

// AddAuthenticator adds an authenticator for a service.
func (m *AuthManager) AddAuthenticator(serviceID string, authenticator Authenticator) error {
	if _, ok := m.authenticators[serviceID]; ok {
		return fmt.Errorf("authenticator for service %s already exists", serviceID)
	}
	m.authenticators[serviceID] = authenticator
	return nil
}

// GetAuthenticator gets the authenticator for a service.
func (m *AuthManager) GetAuthenticator(serviceID string) (Authenticator, bool) {
	authenticator, ok := m.authenticators[serviceID]
	return authenticator, ok
}

// Authenticate authenticates the request for a service.
func (m *AuthManager) Authenticate(ctx context.Context, serviceID string, req *http.Request) (context.Context, error) {
	authenticator, ok := m.GetAuthenticator(serviceID)
	if !ok {
		return ctx, nil
	}
	return authenticator.Authenticate(ctx, req)
}

// SetAPIKey sets the API key for the server.
func (m *AuthManager) SetAPIKey(apiKey string) {
	m.apiKey = apiKey
}

// APIKey returns the API key for the server.
func (m *AuthManager) APIKey() string {
	return m.apiKey
}

// AddOAuth2Authenticator adds an OAuth2 authenticator for a service.
func (m *AuthManager) AddOAuth2Authenticator(ctx context.Context, serviceID string, config *OAuth2Config) error {
	// TODO: Implement this method.
	return nil
}

// RemoveAuthenticator removes the authenticator for a service.
func (m *AuthManager) RemoveAuthenticator(serviceID string) {
	delete(m.authenticators, serviceID)
}
