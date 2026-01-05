// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

type JwksSigner struct {
	key    *rsa.PrivateKey
	keyID  string
	jwk    jose.JSONWebKey
	jwkSet jose.JSONWebKeySet
}

func NewJwksSigner() (*JwksSigner, error) {
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

	return &JwksSigner{
		key:   privateKey,
		keyID: keyID,
		jwk:   jwk,
		jwkSet: jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{jwk},
		},
	}, nil
}

func (s *JwksSigner) NewJWT(issuer string, audience []string) (string, error) {
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

func (s *JwksSigner) Jwks() *jose.JSONWebKeySet {
	return &s.jwkSet
}

func (s *JwksSigner) Key() *rsa.PrivateKey {
	return s.key
}
