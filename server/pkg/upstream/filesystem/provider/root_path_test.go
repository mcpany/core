package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_RootSlashBug(t *testing.T) {
	// Root mapping: / -> /
	rootPaths := map[string]string{"/": "/"}

	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	// Try to access /etc/passwd (assuming it exists on linux/mac)
	// Or use a temporary file for portability
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "mcpany_test_file")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// Access via absolute path since root is /
	resolved, err := p.ResolvePath(tmpFile)

	if err != nil {
		t.Logf("Access denied: %v", err)
	}

	// Expect success now that bug is fixed
	assert.NoError(t, err)

	// Verify resolved path is correct (EvalSymlinks might canonicalize /tmp to /private/tmp on Mac)
	// But it should at least exist
	if err == nil {
		assert.NotEmpty(t, resolved)
	}
}
