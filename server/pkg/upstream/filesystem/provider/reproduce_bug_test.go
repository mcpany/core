// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"
	"strings"

	"github.com/stretchr/testify/require"
)

func TestBug_BrokenSymlinkTraversal(t *testing.T) {
	// 1. Create root dir
	rootDir := t.TempDir()

	// 2. Create outside dir
	outsideDir := t.TempDir()

	// 3. Define a path to a non-existent file in outside dir
	outsideFile := filepath.Join(outsideDir, "ghost.txt")

	// 4. Create a symlink in root pointing to the outside non-existent file
	symlinkPath := filepath.Join(rootDir, "badlink")
	err := os.Symlink(outsideFile, symlinkPath)
	require.NoError(t, err)

	// 5. Setup provider
	p := NewLocalProvider(nil, map[string]string{"/": rootDir}, nil, nil, 0)

	// 6. Try to resolve the symlink path
	resolved, err := p.ResolvePath("/badlink")

	// If the bug exists, this will succeed and return the path to the symlink
	// But it SHOULD fail or return the resolved path (which is outside)

	t.Logf("Resolved: %s", resolved)
	t.Logf("Error: %v", err)

	// Check if the resolved path is safe.
	// The provider claims 'resolved' is safe to use.
	// If we write to 'resolved', do we write to outsideFile?

	if err == nil {
		// Verify where it points
		// If resolved is symlinkPath, writing to it writes to outsideFile
		if resolved == symlinkPath {
			// This confirms we are allowed to access the symlink
			// Now verify that writing to it actually writes outside
			err = os.WriteFile(resolved, []byte("pwned"), 0644)
			if err == nil {
				// Check if outsideFile was created
				if _, err := os.Stat(outsideFile); err == nil {
					t.Fatal("Security Bypass: Successfully wrote to file outside root via broken symlink")
				}
			}
		} else {
			// If it resolved to outsideFile, then checkPathSecurity should have caught it?
			// But wait, if resolved IS outsideFile, then checkPathSecurity would see it is not in rootDir.
			// So if we got here with err==nil, it means resolved WAS considered inside rootDir.
			// Which means resolved probably equals symlinkPath.

			// Let's check if resolved is within rootDir
			rel, err := filepath.Rel(rootDir, resolved)
			if err != nil || strings.HasPrefix(rel, "..") {
				t.Fatalf("Resolved path is outside root: %s", resolved)
			}
		}
	} else {
		t.Log("Great! The provider rejected the broken symlink traversal.")
	}
}
