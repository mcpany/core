package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOCTOU_Vulnerability_Mitigation(t *testing.T) {
	// This test verifies that the safeWriteFile function prevents writing to a file
	// if the path traverses a symlink that was introduced after initial path resolution.

	// Setup filesystem
	baseDir := t.TempDir()
	safeDir := filepath.Join(baseDir, "safe")
	sensitiveDir := filepath.Join(baseDir, "sensitive")
	os.Mkdir(safeDir, 0755)
	os.Mkdir(sensitiveDir, 0755)

	sensitiveFile := filepath.Join(sensitiveDir, "passwd")
	os.WriteFile(sensitiveFile, []byte("secret"), 0600)

	// Simulate the tool logic
    // 1. ResolvePath (Simulated)
    subdir := filepath.Join(safeDir, "subdir")
    os.Mkdir(subdir, 0755)

    // The path we intend to write to (checked and approved)
    targetPath := filepath.Join(subdir, "passwd")

    // 2. Race Condition (Attacker Action)
    // Attacker replaces "subdir" with symlink to "sensitiveDir"
    os.RemoveAll(subdir)
    err := os.Symlink(sensitiveDir, subdir)
    require.NoError(t, err)

    // 3. safeWriteFile (Remediation)
    fs := afero.NewOsFs()
    err = safeWriteFile(fs, targetPath, []byte("pwned"), 0644)

    // Expect error
    assert.Error(t, err, "safeWriteFile should fail if path component became a symlink")

    // 4. Verify No Overwrite
    content, _ := os.ReadFile(sensitiveFile)
    assert.Equal(t, "secret", string(content), "Sensitive file should NOT be overwritten")
}
