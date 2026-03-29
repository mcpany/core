package interop

import (
	"crypto/sha256"
	"fmt"
)

// CostAttestation represents a hardware-attested record of token and compute usage.
//
// Intent: Cryptographically attributes token and compute usage to a specific sub-process lineage,
// ensuring economic accountability across the mesh.
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
type CostAttestation struct {
	LineageID      string `json:"lineage_id"`
	TokensConsumed int    `json:"tokens_consumed"`
	ComputeMs      int    `json:"compute_ms"`
	Signature      string `json:"signature"`
}

// GenerateCostAttestation creates a new CostAttestation with a simulated hardware signature.
//
// Intent: Generates a verifiable record of resource usage for a given task lineage.
//
// Parameters:
//   - lineageID (string): The identifier of the sub-process lineage.
//   - tokens (int): The number of tokens consumed during execution.
//   - computeMs (int): The compute time in milliseconds used during execution.
//
// Returns:
//   - *CostAttestation: A pointer to the generated and signed CostAttestation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Performs SHA-256 hashing to generate a deterministic signature.
func GenerateCostAttestation(lineageID string, tokens int, computeMs int) *CostAttestation {
	// Simulate a hardware-bound (TPM) signature using SHA-256
	data := fmt.Sprintf("%s:%d:%d:haca_secret_key_v1", lineageID, tokens, computeMs)
	hash := sha256.Sum256([]byte(data))
	signature := fmt.Sprintf("%x", hash)

	return &CostAttestation{
		LineageID:      lineageID,
		TokensConsumed: tokens,
		ComputeMs:      computeMs,
		Signature:      signature,
	}
}
