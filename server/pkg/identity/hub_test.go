// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"testing"
	"time"
)

func TestHub_IssueAndVerifyToken(t *testing.T) {
	key := []byte("super-secret-key-12345")
	hub, err := NewHub(key)
	if err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	req := TokenRequest{
		AgentID:     "agent-alpha-001",
		Framework:   "OpenClaw",
		MissionRoot: "mission-root-123",
		TTL:         1 * time.Hour,
	}

	ctx := context.Background()
	token, err := hub.IssueToken(ctx, req)
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	if token == "" {
		t.Fatal("Token is empty")
	}

	claims, err := hub.VerifyToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.AgentID != req.AgentID {
		t.Errorf("Expected AgentID %s, got %s", req.AgentID, claims.AgentID)
	}
	if claims.Framework != req.Framework {
		t.Errorf("Expected Framework %s, got %s", req.Framework, claims.Framework)
	}
	if claims.MissionRoot != req.MissionRoot {
		t.Errorf("Expected MissionRoot %s, got %s", req.MissionRoot, claims.MissionRoot)
	}
	if claims.Issuer != "mcp-any-ztaih" {
		t.Errorf("Expected Issuer mcp-any-ztaih, got %s", claims.Issuer)
	}
}

func TestHub_InvalidToken(t *testing.T) {
	key := []byte("super-secret-key-12345")
	hub, err := NewHub(key)
	if err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	ctx := context.Background()
	_, err = hub.VerifyToken(ctx, "invalid.token.string")
	if err == nil {
		t.Fatal("Expected error verifying invalid token")
	}
}

func TestHub_ExpiredToken(t *testing.T) {
	key := []byte("super-secret-key-12345")
	hub, err := NewHub(key)
	if err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	req := TokenRequest{
		AgentID:     "agent-alpha-001",
		Framework:   "OpenClaw",
		MissionRoot: "mission-root-123",
		TTL:         -1 * time.Hour, // Create an already expired token
	}

	ctx := context.Background()
	token, err := hub.IssueToken(ctx, req)
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	_, err = hub.VerifyToken(ctx, token)
	if err == nil {
		t.Fatal("Expected error verifying expired token")
	}
}

func TestHub_MissingKey(t *testing.T) {
	_, err := NewHub(nil)
	if err == nil {
		t.Fatal("Expected error creating hub with missing key")
	}
	if err != ErrMissingKey {
		t.Errorf("Expected ErrMissingKey, got %v", err)
	}
}
