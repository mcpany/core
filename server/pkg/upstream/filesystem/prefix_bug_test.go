package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLocalPath_PrefixBug(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_prefix_bug")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dataDir := filepath.Join(tempDir, "data")
	err = os.Mkdir(dataDir, 0755)
	require.NoError(t, err)

	// Create a file "foo" inside data directory
	err = os.WriteFile(filepath.Join(dataDir, "foo"), []byte("content"), 0644)
	require.NoError(t, err)

	u := NewUpstream().(*Upstream)

	// Map /data to dataDir
	rootPaths := map[string]string{
		"/data": dataDir,
	}

	// Try to access /datafoo which matches prefix /data but shouldn't match as directory
	resolved, err := u.validateLocalPath("/datafoo", rootPaths)

	// We expect this to fail or NOT resolve to .../data/foo
	// If it returns .../data/foo, then we have confirmed the bug.
	if err == nil {
		// If it resolved successfully, check if it points to the file we created
		expectedPath := filepath.Join(dataDir, "foo")
		// Resolving symlinks/abs path
		expectedPath, _ = filepath.EvalSymlinks(expectedPath)
		expectedPath, _ = filepath.Abs(expectedPath)

		if resolved == expectedPath {
			t.Fatal("Bug confirmed: /datafoo should not resolve to /data/foo")
		}
	}
}

func TestValidateLocalPath_RootRegression(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_root_regression")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file "bar" in tempDir
	err = os.WriteFile(filepath.Join(tempDir, "bar"), []byte("bar content"), 0644)
	require.NoError(t, err)

	u := NewUpstream().(*Upstream)

	// Map / to tempDir
	rootPaths := map[string]string{
		"/": tempDir,
	}

	// Try to access /bar
	resolved, err := u.validateLocalPath("/bar", rootPaths)
	require.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "bar")
	expectedPath, _ = filepath.EvalSymlinks(expectedPath)
	expectedPath, _ = filepath.Abs(expectedPath)

	if resolved != expectedPath {
		t.Fatalf("Regression confirmed: /bar should resolve to %s, got %s", expectedPath, resolved)
	}
}
