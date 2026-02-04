// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill //nolint:revive

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/validation"
	"gopkg.in/yaml.v3"
)

const (
	// SkillFileName is the name of the main skill file.
	SkillFileName = "SKILL.md"
)

var (
	// validNameRegex enforces the naming constraints from the spec.
	// 1-64 chars, lowercase alphanumeric and hyphens. No start/end hyphen. No consecutive hyphens.
	validNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// Manager handles the storage and retrieval of skills.
//
// Summary: Manages the lifecycle and persistence of Skills.
type Manager struct {
	rootDir string
	mu      sync.RWMutex
}

// NewManager creates a new Skill Manager.
//
// Summary: Initializes a new Manager.
//
// Parameters:
//   - rootDir: string. The directory where skills are stored.
//
// Returns:
//   - *Manager: The initialized manager.
//   - error: An error if the root directory cannot be created.
func NewManager(rootDir string) (*Manager, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil { //nolint:gosec
		return nil, fmt.Errorf("failed to create skill root directory: %w", err)
	}
	return &Manager{
		rootDir: rootDir,
	}, nil
}

// ListSkills returns all available skills.
// It scans the root directory for subdirectories containing SKILL.md.
//
// Summary: Lists all skills found in the storage.
//
// Returns:
//   - []*Skill: A list of loaded skills.
//   - error: An error if listing fails.
func (m *Manager) ListSkills() ([]*Skill, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill root directory: %w", err)
	}

	skills := make([]*Skill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skill, err := m.loadSkill(entry.Name())
		if err != nil {
			logging.GetLogger().Warn("Failed to load skill", "name", entry.Name(), "error", err)
			continue
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

// GetSkill retrieves a specific skill by name.
//
// Summary: Retrieves a skill by name.
//
// Parameters:
//   - name: string. The name of the skill.
//
// Returns:
//   - *Skill: The loaded skill.
//   - error: An error if the skill is not found or fails to load.
func (m *Manager) GetSkill(name string) (*Skill, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loadSkill(name)
}

// CreateSkill creates a new skill.
// It ensures the name is valid and the directory doesn't already exist.
//
// Summary: Creates a new skill.
//
// Parameters:
//   - skill: *Skill. The skill definition to create.
//
// Returns:
//   - error: An error if validation fails or creation fails.
func (m *Manager) CreateSkill(skill *Skill) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := validateName(skill.Name); err != nil {
		return err
	}

	skillDir := filepath.Join(m.rootDir, skill.Name)
	if _, err := os.Stat(skillDir); err == nil {
		return fmt.Errorf("skill already exists: %s", skill.Name)
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil { //nolint:gosec
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Create optional directories
	for _, dir := range []string{"scripts", "references", "assets"} {
		_ = os.Mkdir(filepath.Join(skillDir, dir), 0755) //nolint:gosec
	}

	return m.writeSkillFile(skillDir, skill)
}

// UpdateSkill updates an existing skill.
// If the name has changed, it renames the directory.
//
// Summary: Updates an existing skill.
//
// Parameters:
//   - originalName: string. The current name of the skill.
//   - skill: *Skill. The updated skill definition.
//
// Returns:
//   - error: An error if the skill is not found, invalid, or update fails.
func (m *Manager) UpdateSkill(originalName string, skill *Skill) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := validateName(skill.Name); err != nil {
		return err
	}

	originalDir := filepath.Join(m.rootDir, originalName)
	newDir := filepath.Join(m.rootDir, skill.Name)

	if _, err := os.Stat(originalDir); os.IsNotExist(err) {
		return fmt.Errorf("skill not found: %s", originalName)
	}

	// Handle rename
	if originalName != skill.Name {
		if _, err := os.Stat(newDir); err == nil {
			return fmt.Errorf("destination skill already exists: %s", skill.Name)
		}
		if err := os.Rename(originalDir, newDir); err != nil {
			return fmt.Errorf("failed to rename skill: %w", err)
		}
	}

	return m.writeSkillFile(newDir, skill)
}

// DeleteSkill deletes a skill.
//
// Summary: Deletes a skill.
//
// Parameters:
//   - name: string. The name of the skill to delete.
//
// Returns:
//   - error: An error if the skill is not found or deletion fails.
func (m *Manager) DeleteSkill(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skillDir := filepath.Join(m.rootDir, name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill not found: %s", name)
	}

	// Remove the entire directory
	return os.RemoveAll(skillDir)
}

// SaveAsset saves an asset file (script, reference, etc.) for a skill.
//
// Summary: Saves an asset file within a skill directory.
//
// Parameters:
//   - skillName: string. The name of the skill.
//   - relPath: string. The relative path of the asset within the skill directory.
//   - content: []byte. The content of the asset file.
//
// Returns:
//   - error: An error if validation fails or writing fails.
func (m *Manager) SaveAsset(skillName string, relPath string, content []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// validate path to prevent traversal and ensure it is relative
	if err := validation.IsSecureRelativePath(relPath); err != nil {
		return fmt.Errorf("invalid asset path: %w", err)
	}

	cleanPath := filepath.Clean(relPath)
	skillDir := filepath.Join(m.rootDir, skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill not found: %s", skillName)
	}

	fullPath := filepath.Join(skillDir, cleanPath)

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil { //nolint:gosec
		return fmt.Errorf("failed to create asset parent directory: %w", err)
	}

	return os.WriteFile(fullPath, content, 0644) //nolint:gosec
}

func (m *Manager) loadSkill(name string) (*Skill, error) {
	skillDir := filepath.Join(m.rootDir, name)
	content, err := os.ReadFile(filepath.Join(skillDir, SkillFileName)) //nolint:gosec
	if err != nil {
		return nil, err
	}

	var skill Skill
	skill.Path = skillDir
	skill.Name = name // Use directory name as source of truth for ID/Name context

	// Parse Frontmatter + Body
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) >= 3 && parts[0] == "" {
		// Valid frontmatter
		if err := yaml.Unmarshal([]byte(parts[1]), &skill.Frontmatter); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
		}
		skill.Instructions = strings.TrimSpace(parts[2])
	} else {
		// No frontmatter? or malformed. Spec requires frontmatter.
		// We'll treat it as error or just body? Spec says "must contain".
		return nil, fmt.Errorf("invalid SKILL.md format (missing frontmatter)")
	}

	// Validate name consistency (optional, but good practice)
	// if skill.Name != name {
	// 	// Warn? or Override? Directory name usually rules in filesystem based systems.
	// 	// Let's just note it.
	// }

	// List assets
	_ = filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() != SkillFileName {
			rel, _ := filepath.Rel(skillDir, path)
			skill.Assets = append(skill.Assets, rel)
		}
		return nil
	})

	return &skill, nil
}

func (m *Manager) writeSkillFile(dir string, skill *Skill) error {
	// Marshal frontmatter
	fmData, err := yaml.Marshal(skill.Frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n%s", string(fmData), skill.Instructions)
	return os.WriteFile(filepath.Join(dir, SkillFileName), []byte(content), 0644) //nolint:gosec
}

func validateName(name string) error {
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("invalid skill name %q: must be 1-64 chars, lowercase alphanumeric and hyphens only", name)
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("invalid skill name %q: cannot contain consecutive hyphens", name)
	}
	return nil
}
