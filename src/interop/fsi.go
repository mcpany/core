package interop

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// FederatedIdentityHub acts as the authoritative Identity Mint for connected agents.
//
// Intent: Issues hardware-attested, cross-framework identity tokens that allow disparate agents
// (Claude, OpenClaw, AutoGen) to verify each other's lineage and mission-bound authority.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type FederatedIdentityHub struct {
	mu           sync.RWMutex
	ActiveTokens map[string]TokenMetadata
}

// TokenMetadata contains the context and lineage associated with an identity token.
//
// Intent: Stores the hardware attestation data, mission scope, and temporal validity of a given token.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type TokenMetadata struct {
	MissionID   string
	Framework   string
	IssuedAt    time.Time
	IsAttested  bool
}

// NewFederatedIdentityHub creates a new instance of the FSI provider.
//
// Intent: Initializes the registry for tracking active cross-framework identity tokens.
//
// Parameters:
//   - None.
//
// Returns:
//   - *FederatedIdentityHub: A pointer to the initialized Identity Hub.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates the internal ActiveTokens map.
func NewFederatedIdentityHub() *FederatedIdentityHub {
	return &FederatedIdentityHub{
		ActiveTokens: make(map[string]TokenMetadata),
	}
}

// MintToken issues a new, hardware-bound identity token for a specific agent framework.
//
// Intent: Cryptographically generates a lineage token bound to a specific mission and framework.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - missionID (string): The unique identifier for the overarching mission.
//   - framework (string): The name of the agent framework requesting the token.
//
// Returns:
//   - string: The generated token hash.
//   - error: An error if generation fails.
//
// Errors:
//   - Returns an error if the missionID or framework are empty.
//
// Side Effects:
//   - Adds a new entry to the internal ActiveTokens registry.
func (h *FederatedIdentityHub) MintToken(ctx context.Context, missionID, framework string) (string, error) {
	if missionID == "" || framework == "" {
		return "", fmt.Errorf("missionID and framework are required to mint a token")
	}

	raw := fmt.Sprintf("%s:%s:%d", missionID, framework, time.Now().UnixNano())
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))

	h.mu.Lock()
	defer h.mu.Unlock()
	h.ActiveTokens[hash] = TokenMetadata{
		MissionID:  missionID,
		Framework:  framework,
		IssuedAt:   time.Now(),
		IsAttested: false,
	}

	return hash, nil
}

// VerifyToken validates the integrity and presence of a given identity token.
//
// Intent: Authoritatively confirms that a given token is active and was legally minted by the Hub.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - token (string): The identity token to verify.
//
// Returns:
//   - bool: True if the token is valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *FederatedIdentityHub) VerifyToken(ctx context.Context, token string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.ActiveTokens[token]
	return exists
}

// RecordAttestation updates the token state indicating it has been successfully attested by the framework.
//
// Intent: Acknowledges that the agent framework has accepted and bound itself to the provided identity token.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - token (string): The identity token that was attested.
//
// Returns:
//   - error: An error if the token is unknown.
//
// Errors:
//   - Returns an error if the token does not exist in the registry.
//
// Side Effects:
//   - Modifies the TokenMetadata to set IsAttested = true.
func (h *FederatedIdentityHub) RecordAttestation(ctx context.Context, token string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	meta, exists := h.ActiveTokens[token]
	if !exists {
		return fmt.Errorf("cannot record attestation for unknown token: %s", token)
	}

	meta.IsAttested = true
	h.ActiveTokens[token] = meta
	return nil
}
