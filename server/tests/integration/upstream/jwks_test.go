// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
