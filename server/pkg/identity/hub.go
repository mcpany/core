// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrMissingKey   = errors.New("missing signing key")
)

// TokenRequest represents the parameters for issuing a new mesh-resident token.
type TokenRequest struct {
	AgentID     string
	Framework   string
	MissionRoot string
	TTL         time.Duration
}

// AgentClaims represents the custom claims embedded in the agent's JWT.
type AgentClaims struct {
	AgentID     string `json:"agent_id"`
	Framework   string `json:"framework"`
	MissionRoot string `json:"mission_root"`
	jwt.RegisteredClaims
}

// IdentityManager defines the interface for the Zero-Trust Agent Identity Hub.
type IdentityManager interface {
	IssueToken(ctx context.Context, req TokenRequest) (string, error)
	VerifyToken(ctx context.Context, tokenString string) (*AgentClaims, error)
}

// Hub is the authoritative local identity service for agents.
type Hub struct {
	signingKey []byte
}

// NewHub creates a new Zero-Trust Agent Identity Hub.
func NewHub(signingKey []byte) (*Hub, error) {
	if len(signingKey) == 0 {
		return nil, ErrMissingKey
	}
	return &Hub{signingKey: signingKey}, nil
}

// IssueToken mints a new mesh-resident token for an agent.
func (h *Hub) IssueToken(ctx context.Context, req TokenRequest) (string, error) {
	if req.TTL == 0 {
		req.TTL = time.Hour // Default to 1 hour
	}

	claims := AgentClaims{
		AgentID:     req.AgentID,
		Framework:   req.Framework,
		MissionRoot: req.MissionRoot,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(req.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mcp-any-ztaih",
			Subject:   req.AgentID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.signingKey)
}

// VerifyToken verifies the authenticity and validity of a mesh-resident token.
func (h *Hub) VerifyToken(ctx context.Context, tokenString string) (*AgentClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AgentClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.signingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*AgentClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
