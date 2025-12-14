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

package upstream

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwksSigner(t *testing.T) {
	t.Run("newJwksSigner", func(t *testing.T) {
		signer, err := newJwksSigner()
		require.NoError(t, err)
		assert.NotNil(t, signer.key)
		assert.NotEmpty(t, signer.keyID)
		assert.NotNil(t, signer.jwk)
		assert.NotNil(t, signer.jwkSet)
		assert.Len(t, signer.jwkSet.Keys, 1)
	})

	t.Run("newJWT", func(t *testing.T) {
		signer, err := newJwksSigner()
		require.NoError(t, err)

		issuer := "test-issuer"
		audience := []string{"test-audience"}
		tokenString, err := signer.newJWT(issuer, audience)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (interface{}, error) {
			return signer.key.Public(), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		require.True(t, ok)
		assert.Equal(t, issuer, claims.Issuer)
		assert.Equal(t, audience, []string(claims.Audience))

		// Check expiry and issued at claims
		require.NotNil(t, claims.ExpiresAt)
		assert.True(t, time.Now().Before(claims.ExpiresAt.Time))

		require.NotNil(t, claims.IssuedAt)
		assert.True(t, time.Now().After(claims.IssuedAt.Time))
	})

	t.Run("jwks", func(t *testing.T) {
		signer, err := newJwksSigner()
		require.NoError(t, err)

		jwks := signer.jwks()
		assert.NotNil(t, jwks)
		assert.Len(t, jwks.Keys, 1)
		assert.Equal(t, signer.jwk, jwks.Keys[0])
	})
}
