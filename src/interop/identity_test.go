package interop_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/src/interop"
)

// TestZeroTrustIdentityHub validates the minting, verification, and revocation
// of hardware-attested, mesh-resident identity tokens.
func TestZeroTrustIdentityHub(t *testing.T) {
	ctx := context.Background()
	hub := interop.NewZeroTrustIdentityHub()

	t.Run("MintToken", func(t *testing.T) {
		token, err := hub.MintToken(ctx, "OpenClaw", "mission_root_xyz", []string{"adaptive_reasoning", "multi_agent_chat"})
		if err != nil {
			t.Fatalf("Failed to mint token: %v", err)
		}

		if token.TokenID == "" {
			t.Error("Minted token should have a non-empty TokenID")
		}

		if token.Framework != "OpenClaw" {
			t.Errorf("Expected framework OpenClaw, got %s", token.Framework)
		}

		if token.MissionRoot != "mission_root_xyz" {
			t.Errorf("Expected mission root 'mission_root_xyz', got %s", token.MissionRoot)
		}

		if token.Revoked {
			t.Error("Newly minted token should not be revoked")
		}

		if time.Since(token.IssuedAt) > time.Second {
			t.Errorf("IssuedAt timestamp seems wrong: %v", token.IssuedAt)
		}
	})

	t.Run("VerifyToken_Success", func(t *testing.T) {
		token, _ := hub.MintToken(ctx, "CrewAI", "mission_root_cai", []string{"task_delegation"})

		valid, err := hub.VerifyToken(ctx, token.TokenID, "task_delegation")
		if err != nil {
			t.Fatalf("Failed to verify token: %v", err)
		}

		if !valid {
			t.Error("Expected token to be valid for the given capability")
		}
	})

	t.Run("VerifyToken_MissingCapability", func(t *testing.T) {
		token, _ := hub.MintToken(ctx, "CrewAI", "mission_root_cai", []string{"role_discovery"})

		valid, err := hub.VerifyToken(ctx, token.TokenID, "task_delegation")
		if err == nil {
			t.Error("Expected error when verifying missing capability, got nil")
		}

		if valid {
			t.Error("Expected token verification to return false")
		}
	})

	t.Run("VerifyToken_NonExistent", func(t *testing.T) {
		valid, err := hub.VerifyToken(ctx, "fake_token_id", "some_cap")
		if err == nil {
			t.Error("Expected error for non-existent token, got nil")
		}

		if valid {
			t.Error("Expected token verification to return false")
		}
	})

	t.Run("RevokeToken", func(t *testing.T) {
		token, _ := hub.MintToken(ctx, "AutoGen", "mission_root_ag", []string{"subagent_exec"})

		err := hub.RevokeToken(ctx, token.TokenID)
		if err != nil {
			t.Fatalf("Failed to revoke token: %v", err)
		}

		// Verify should now fail
		valid, err := hub.VerifyToken(ctx, token.TokenID, "subagent_exec")
		if err == nil {
			t.Error("Expected error when verifying revoked token, got nil")
		}
		if err.Error() != "invalid token: revoked" {
			t.Errorf("Expected error 'invalid token: revoked', got %v", err)
		}
		if valid {
			t.Error("Expected token verification to return false")
		}
	})

	t.Run("RevokeToken_NonExistent", func(t *testing.T) {
		err := hub.RevokeToken(ctx, "fake_token_id")
		if err == nil {
			t.Error("Expected error for revoking non-existent token, got nil")
		}
	})
}
