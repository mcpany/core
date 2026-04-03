// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package attestation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNonExistenceProof_Success(t *testing.T) {
	gateway := NewGateway("test-secret-key")

	// Use a file that definitely doesn't exist
	tempDir := t.TempDir()
	missingFile := filepath.Join(tempDir, "does-not-exist.txt")

	proof, err := gateway.GenerateNonExistenceProof(missingFile)
	require.NoError(t, err)
	require.NotNil(t, proof)

	assert.Equal(t, missingFile, proof.FilePath)
	assert.NotEmpty(t, proof.Signature)
	assert.WithinDuration(t, time.Now().UTC(), proof.Timestamp, time.Second)
}

func TestGenerateNonExistenceProof_FailureFileExists(t *testing.T) {
	gateway := NewGateway("test-secret-key")

	// Create a file
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	err := os.WriteFile(existingFile, []byte("content"), 0644)
	require.NoError(t, err)

	proof, err := gateway.GenerateNonExistenceProof(existingFile)
	assert.Error(t, err)
	assert.Nil(t, proof)
	assert.Contains(t, err.Error(), "file exists")
}
