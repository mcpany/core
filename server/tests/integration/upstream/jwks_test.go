// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mcpany/core/server/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwksSigner(t *testing.T) {
	t.Run("newJwksSigner", func(t *testing.T) {
		signer, err := test.NewJwksSigner()
		require.NoError(t, err)
		assert.NotNil(t, signer)
	})

	t.Run("newJWT", func(t *testing.T) {
		signer, err := test.NewJwksSigner()
		require.NoError(t, err)

		issuer := "test-issuer"
		audience := []string{"test-audience"}
		tokenString, err := signer.NewJWT(issuer, audience)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (interface{}, error) {
			return signer.Key().Public(), nil
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
		signer, err := test.NewJwksSigner()
		require.NoError(t, err)

		jwks := signer.Jwks()
		assert.NotNil(t, jwks)
		assert.Len(t, jwks.Keys, 1)
	})
}
