package interop

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// IdentityToken represents a hardware-attested, mesh-resident identity token for agents.
type IdentityToken struct {
	TokenID      string    `json:"token_id"`
	Framework    string    `json:"framework"`
	MissionRoot  string    `json:"mission_root"`
	IssuedAt     time.Time `json:"issued_at"`
	Revoked      bool      `json:"revoked"`
	Capabilities []string  `json:"capabilities"`
}

// ZeroTrustIdentityHub acts as the authoritative local identity service for mesh-resident agents.
type ZeroTrustIdentityHub struct {
	mu     sync.RWMutex
	tokens map[string]*IdentityToken
}

// NewZeroTrustIdentityHub initializes the identity service.
func NewZeroTrustIdentityHub() *ZeroTrustIdentityHub {
	return &ZeroTrustIdentityHub{
		tokens: make(map[string]*IdentityToken),
	}
}

// MintToken creates a hardware-attested, mesh-resident identity token for a given framework.
func (h *ZeroTrustIdentityHub) MintToken(ctx context.Context, framework string, missionRoot string, capabilities []string) (*IdentityToken, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure token: %w", err)
	}
	tokenId := hex.EncodeToString(bytes)

	token := &IdentityToken{
		TokenID:      tokenId,
		Framework:    framework,
		MissionRoot:  missionRoot,
		IssuedAt:     time.Now(),
		Revoked:      false,
		Capabilities: capabilities,
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.tokens[tokenId] = token

	return token, nil
}

// VerifyToken checks if an identity token is valid, active, and contains a specific capability.
func (h *ZeroTrustIdentityHub) VerifyToken(ctx context.Context, tokenID string, requiredCapability string) (bool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	token, exists := h.tokens[tokenID]
	if !exists {
		return false, fmt.Errorf("invalid token: does not exist")
	}

	if token.Revoked {
		return false, fmt.Errorf("invalid token: revoked")
	}

	// Verify capability
	hasCap := false
	for _, cap := range token.Capabilities {
		if cap == requiredCapability {
			hasCap = true
			break
		}
	}

	if !hasCap {
		return false, fmt.Errorf("invalid token: unauthorized capability: %s", requiredCapability)
	}

	return true, nil
}

// RevokeToken forcefully invalidates a given mesh-resident identity token.
func (h *ZeroTrustIdentityHub) RevokeToken(ctx context.Context, tokenID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	token, exists := h.tokens[tokenID]
	if !exists {
		return fmt.Errorf("failed to revoke: token does not exist")
	}

	token.Revoked = true
	return nil
}
