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
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

type jwksSigner struct {
	key    *rsa.PrivateKey
	keyID  string
	jwk    jose.JSONWebKey
	jwkSet jose.JSONWebKeySet
}

func newJwksSigner() (*jwksSigner, error) {
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

func (s *jwksSigner) newJWT(issuer string, audience []string) (string, error) {
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

func (s *jwksSigner) jwks() *jose.JSONWebKeySet {
	return &s.jwkSet
}
