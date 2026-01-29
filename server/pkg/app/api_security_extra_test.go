package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUploadSkillAsset_PathDisclosure(t *testing.T) {
	// Create a directory
	tmpDir := t.TempDir()

	// Create manager with this root
	sm, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	// Create a skill
	skillName := "leak-skill"
	err = sm.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: skillName}, Instructions: "inst"})
	require.NoError(t, err)

	// Now make the skill directory read-only to cause SaveAsset to fail
	skillDir := filepath.Join(tmpDir, skillName)
	// Make it 0500 (read-only, executable for traversal)
	err = os.Chmod(skillDir, 0500)
	require.NoError(t, err)

	// Ensure we can clean up
	t.Cleanup(func() {
		_ = os.Chmod(skillDir, 0755)
	})

	app := &Application{SkillManager: sm}

	// Try to upload an asset
	// Path must match: /api/v1/skills/{name}/assets
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillName+"/assets?path=test.txt", bytes.NewReader([]byte("data")))
	w := httptest.NewRecorder()
	app.handleUploadSkillAsset().ServeHTTP(w, req)

	// We expect 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check for leak
	// The error likely contains the full path if not handled correctly
	if strings.Contains(w.Body.String(), tmpDir) {
		t.Errorf("Security Leak! Response body contains server path: %s", w.Body.String())
	} else {
		t.Logf("Response body is safe: %s", w.Body.String())
	}

	assert.Contains(t, w.Body.String(), "Failed to save asset")
}
