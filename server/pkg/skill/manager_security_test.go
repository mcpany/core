// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-sec-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	require.NoError(t, err)

	skillName := "sec-skill"
	err = m.CreateSkill(&Skill{Frontmatter: Frontmatter{Name: skillName}})
	require.NoError(t, err)

	testCases := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{"ValidFile", "script.py", false},
		{"ValidSubDir", "scripts/script.py", false},
		{"TraversalUp", "../evil.py", true},
		{"TraversalRoot", "/etc/passwd", true}, // Absolute path
		{"TraversalSibling", "foo/../../bar", true},
		{"TraversalDeep", "a/b/c/../../../d", false}, // resolving to a/d which is valid if d is inside skill? Wait. a/b/c/../../../d -> d. Valid.
		{"TraversalEscape", "a/../../../../etc/passwd", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := m.SaveAsset(skillName, tc.path, []byte("data"))
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
