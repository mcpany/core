// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAuditStore_Chain(t *testing.T) {
	// Create a temporary file
	f, err := os.CreateTemp("", "audit_file_chain_*.log")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	// Open store
	store, err := NewFileAuditStore(path)
	require.NoError(t, err)

	// Write entries
	entry1 := AuditEntry{Timestamp: time.Now(), ToolName: "tool1"}
	err = store.Write(context.Background(), entry1)
	require.NoError(t, err)

	entry2 := AuditEntry{Timestamp: time.Now(), ToolName: "tool2"}
	err = store.Write(context.Background(), entry2)
	require.NoError(t, err)

	store.Close()

	// Read back and verify chain
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 2)

	var e1, e2 AuditEntry
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &e1))
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &e2))

	// Verify e1
	// Expected hash calc manually? No, just rely on computeHash being deterministic.
	// But we can check e2.PrevHash == e1.Hash
	assert.NotEmpty(t, e1.Hash)
	assert.Empty(t, e1.PrevHash) // First entry

	assert.NotEmpty(t, e2.Hash)
	assert.Equal(t, e1.Hash, e2.PrevHash, "Chain broken: e2.PrevHash != e1.Hash")
}

func TestFileAuditStore_Restart(t *testing.T) {
	f, err := os.CreateTemp("", "audit_file_restart_*.log")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	// Session 1
	store1, err := NewFileAuditStore(path)
	require.NoError(t, err)

	entry1 := AuditEntry{Timestamp: time.Now(), ToolName: "tool1"}
	require.NoError(t, store1.Write(context.Background(), entry1))
	store1.Close()

	// Read file to get hash of entry1
	content, _ := os.ReadFile(path)
	var e1 AuditEntry
	json.Unmarshal(content, &e1)

	// Session 2 - should pick up last hash
	store2, err := NewFileAuditStore(path)
	require.NoError(t, err)

	// Check internal state (using reflection or just writing next entry)
	// We'll write next entry and check its PrevHash
	entry2 := AuditEntry{Timestamp: time.Now(), ToolName: "tool2"}
	require.NoError(t, store2.Write(context.Background(), entry2))
	store2.Close()

	// Read back
	content, _ = os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 2)

	var e2 AuditEntry
	json.Unmarshal([]byte(lines[1]), &e2)

	assert.Equal(t, e1.Hash, e2.PrevHash, "Restart failed to preserve chain")
}

func TestReadLastHash_LargeLine(t *testing.T) {
	f, err := os.CreateTemp("", "audit_large_*.log")
	require.NoError(t, err)
	path := f.Name()
	defer os.Remove(path)

	// Create a large entry > 4KB
	largeArg := strings.Repeat("a", 5000)
	entry := AuditEntry{
		Timestamp: time.Now(),
		ToolName: "large_tool",
		Arguments: json.RawMessage(`"` + largeArg + `"`),
		Hash: "target_hash",
	}

	// Write it manually
	bytes, _ := json.Marshal(entry)
	f.Write(bytes)
	f.WriteString("\n")
	f.Close()

	// Test readLastHash
	hash, err := readLastHash(path)
	require.NoError(t, err)
	assert.Equal(t, "target_hash", hash)
}

func TestReadLastHash_MultipleLines(t *testing.T) {
	f, err := os.CreateTemp("", "audit_multi_*.log")
	require.NoError(t, err)
	path := f.Name()
	defer os.Remove(path)

	entry1 := AuditEntry{Hash: "hash1"}
	bytes1, _ := json.Marshal(entry1)
	f.Write(bytes1)
	f.WriteString("\n")

	entry2 := AuditEntry{Hash: "hash2"}
	bytes2, _ := json.Marshal(entry2)
	f.Write(bytes2)
	f.WriteString("\n")
	f.Close()

	hash, err := readLastHash(path)
	require.NoError(t, err)
	assert.Equal(t, "hash2", hash)
}
