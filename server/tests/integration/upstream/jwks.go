// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package upstream contains integration tests
package upstream

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

type jwksSigner struct { //nolint:unused
	key    *rsa.PrivateKey
	keyID  string
	jwk    jose.JSONWebKey
	jwkSet jose.JSONWebKeySet
}

func newJwksSigner() (*jwksSigner, error) { //nolint:unused
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	keyID := "test-key"
	jwk := jose.JSONWebKey{
		Key:       privateKey.Public(),
		KeyID:     keyID,
		Algorithm: "RS256",
		Use:       "sig",
	}

	return &jwksSigner{
		key:   privateKey,
		keyID: keyID,
		jwk:   jwk,
		jwkSet: jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{jwk},
		},
	}, nil
}

func (s *jwksSigner) newJWT(issuer string, audience []string) (string, error) { //nolint:unused
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Audience:  audience,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.key)
}

func (s *jwksSigner) jwks() *jose.JSONWebKeySet { //nolint:unused
	return &s.jwkSet
}
