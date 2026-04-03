// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

// NonExistenceProof represents a cryptographically signed proof that a specific file does not exist
// on the filesystem at a given point in time. This mitigates CVE-2026-25725 (config-injection).
type NonExistenceProof struct {
	FilePath  string
	Timestamp time.Time
	Signature string // Simplified for now
}

// Gateway represents the Deterministic Attestation Gateway.
type Gateway struct {
	SigningKey string
}

// NewGateway creates a new Gateway.
func NewGateway(signingKey string) *Gateway {
	return &Gateway{
		SigningKey: signingKey,
	}
}

// GenerateNonExistenceProof generates a proof that a file does NOT exist.
// Returns an error if the file DOES exist or if there's an issue accessing the filesystem.
func (g *Gateway) GenerateNonExistenceProof(filepath string) (*NonExistenceProof, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return nil, fmt.Errorf("file exists: cannot generate non-existence proof for %s", filepath)
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}

	timestamp := time.Now().UTC()

	// Create a simple deterministic signature
	dataToSign := fmt.Sprintf("non_existence:%s:%d:%s", filepath, timestamp.Unix(), g.SigningKey)
	hash := sha256.Sum256([]byte(dataToSign))
	signature := hex.EncodeToString(hash[:])

	return &NonExistenceProof{
		FilePath:  filepath,
		Timestamp: timestamp,
		Signature: signature,
	}, nil
}
