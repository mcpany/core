/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Manager handles the authentication and token validation logic.
type Manager struct {
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	issuerURL  string
	audience   string
	resource   string
}

// NewAuthManager creates a new authentication manager.
func NewAuthManager(issuerURL, audience, resource string) *Manager {
	if issuerURL == "" {
		return &Manager{}
	}

	provider, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		panic(fmt.Sprintf("failed to create OIDC provider: %v", err))
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: audience,
	})

	return &Manager{
		provider:   provider,
		verifier:   verifier,
		issuerURL:  issuerURL,
		audience:   audience,
		resource:   resource,
	}
}

// VerifyToken verifies the provided ID token.
func (m *Manager) VerifyToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	if m.verifier == nil {
		return nil, fmt.Errorf("auth manager not configured")
	}
	return m.verifier.Verify(ctx, rawIDToken)
}

// IsEnabled returns true if the authentication manager is configured.
func (m *Manager) IsEnabled() bool {
	return m.issuerURL != ""
}

// ProtectedResourceMetadataHandler returns an HTTP handler that serves the
// .well-known/oauth-protected-resource metadata.
func (m *Manager) ProtectedResourceMetadataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.IsEnabled() {
			http.NotFound(w, r)
			return
		}

		metadata := struct {
			Resource              string   `json:"resource"`
			AuthorizationServers []string `json:"authorization_servers"`
			ScopesSupported      []string `json:"scopes_supported"`
		}{
			Resource:              m.resource,
			AuthorizationServers: []string{m.issuerURL},
			ScopesSupported:      []string{"mcp:tools", "mcp:resources"},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
