package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAsset_PathTraversal_Repro(t *testing.T) {
	// Create a temporary directory for the skills
	tmpDir := t.TempDir()

	// Create a subdirectory inside tmpDir to serve as the skills root
	// This ensures that ".." points to tmpDir which is writable and controlled by us.
	skillsRoot := filepath.Join(tmpDir, "skills_root")
	if err := os.Mkdir(skillsRoot, 0755); err != nil {
		t.Fatalf("Failed to create skills root: %v", err)
	}

	manager, err := NewManager(skillsRoot)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Attack Payload: use ".." as skill name to target the parent directory (tmpDir)
	skillName := ".."
	assetName := "pwned.txt"
	content := []byte("pwned")

	// Attempt to save the asset using the malicious skill name
	err = manager.SaveAsset(skillName, assetName, content)

	// We expect this to FAIL with the fix.
	if err == nil {
		// Verify if the file was written to tmpDir/pwned.txt
		pwnedPath := filepath.Join(tmpDir, assetName)
		if _, statErr := os.Stat(pwnedPath); statErr == nil {
			t.Errorf("VULNERABILITY CONFIRMED: Wrote file to %s using skillName='..'", pwnedPath)
		} else {
			// Even if file not found, no error means something went wrong with validation
			t.Errorf("SaveAsset succeeded unexpectedly (though file not found at %s).", pwnedPath)
		}
	} else {
		// Expected error
		t.Logf("SaveAsset failed as expected: %v", err)
	}
}

func TestSaveAsset_OverwriteSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)

	skillName := "test-skill"
	// Initialize with Frontmatter
	skill := &Skill{
		Frontmatter: Frontmatter{
			Name: skillName,
			Description: "Test skill",
		},
		Instructions: "Original instructions",
	}

	if err := manager.CreateSkill(skill); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Attempt to overwrite SKILL.md
	assetContent := []byte("Hacked instructions")
	err := manager.SaveAsset(skillName, SkillFileName, assetContent)

	if err == nil {
		t.Errorf("Expected error when overwriting %s, but got nil", SkillFileName)
	} else {
		t.Logf("Overwrite prevented as expected: %v", err)
	}
}

func TestSaveAsset_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)
	skillName := "test-skill-valid"

	skill := &Skill{
		Frontmatter: Frontmatter{Name: skillName},
		Instructions: "Instructions",
	}
	if err := manager.CreateSkill(skill); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	err := manager.SaveAsset(skillName, "scripts/test.py", []byte("print('hello')"))
	if err != nil {
		t.Errorf("SaveAsset failed for valid file: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, skillName, "scripts/test.py")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("File not created at %s", path)
	}
}
