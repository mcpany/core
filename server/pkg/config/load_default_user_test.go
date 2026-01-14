package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultUser_ShouldNotAccessDisabledProfiles(t *testing.T) {
	content := `
global_settings: {
    profiles: ["enabled_profile"]
    profile_definitions: [
        { name: "enabled_profile" },
        { name: "disabled_profile" }
    ]
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.textproto")
	fs := afero.NewOsFs()
	f, err := fs.Create(filePath)
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	fileStore := NewFileStore(fs, []string{filePath})
	cfg, err := LoadServices(context.Background(), fileStore, "server")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotEmpty(t, cfg.GetUsers())
	defaultUser := cfg.GetUsers()[0]

	// Verify the default user has the enabled profile
	assert.Contains(t, defaultUser.GetProfileIds(), "enabled_profile")

	// Verify the default user DOES NOT have the disabled profile
	// This assertion is expected to fail if the bug exists
	assert.NotContains(t, defaultUser.GetProfileIds(), "disabled_profile", "Default user should not have access to disabled profiles")
}
