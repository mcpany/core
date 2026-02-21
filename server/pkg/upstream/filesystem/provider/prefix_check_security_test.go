package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWeakPrefixCheckRepro(t *testing.T) {
	// Setup temporary directory structure
	tmpDir := t.TempDir()

	// Create "safe" directory
	safeDir := filepath.Join(tmpDir, "safe")
	err := os.Mkdir(safeDir, 0755)
	require.NoError(t, err)

	// Create "safe_but_secret" directory (sibling, shares prefix)
	secretDir := filepath.Join(tmpDir, "safe_but_secret")
	err = os.Mkdir(secretDir, 0755)
	require.NoError(t, err)

	// Create a secret file in secret directory
	secretFile := filepath.Join(secretDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("s3cr3t"), 0644)
	require.NoError(t, err)

	// Configure provider allowing "safe" directory
	// We map "/" to tmpDir so we can access everything under tmpDir potentially
	rootPaths := map[string]string{
		"/": tmpDir,
	}

	// allowedPaths includes ONLY safeDir
	allowedPaths := []string{safeDir}

	// Use 0 for symlinkMode (default)
	provider := NewLocalProvider(nil, rootPaths, allowedPaths, nil, 0)

	// Try to resolve path to secret file
	// virtual path: /safe_but_secret/secret.txt (maps to secretFile)
	// allowed path: .../safe
	// target path: .../safe_but_secret/secret.txt
	// prefix check: HasPrefix(".../safe_but_secret/secret.txt", ".../safe") is TRUE!

	// Determine virtual path relative to root
	relSecret, err := filepath.Rel(tmpDir, secretFile)
	require.NoError(t, err)
	virtualPath := "/" + relSecret

	resolved, err := provider.ResolvePath(virtualPath)

	// If err is nil, it means access was allowed -> Vulnerability!
	if err == nil {
		t.Logf("VULNERABILITY CONFIRMED: Accessed %s which is outside allowed path %s", resolved, safeDir)
		t.Fail()
	} else {
		t.Logf("Blocked as expected: %v", err)
	}
}
