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
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock IDTokenVerifier for testing
type mockIDTokenVerifier struct {
	verify func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

func (m *mockIDTokenVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return m.verify(ctx, rawIDToken)
}

func TestOAuth2Authenticator_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("successful authentication", func(t *testing.T) {
		verifier := &mockIDTokenVerifier{
			verify: func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
				// Create a dummy IDToken for testing
				claims := []byte(`{"email": "test@example.com"}`)
				token := &oidc.IDToken{
					Issuer:  "https://example.com",
					Audience: []string{"test-client"},
					Subject: "test-subject",
					Expiry:   time.Now().Add(time.Hour),
					IssuedAt: time.Now(),
					Nonce:   "test-nonce",
				}
				// Unmarshal claims into the token
				if err := token.Claims(&claims); err != nil {
					return nil, err
				}
				return token, nil
			},
		}

		authenticator := &OAuth2Authenticator{verifier: verifier}
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		newCtx, err := authenticator.Authenticate(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", newCtx.Value("user"))
	})
}
