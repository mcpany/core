// Copyright 2026 Author(s) of MCP Any
// Frontmatter represents the YAML frontmatter of a SKILL.md file.
//
// Summary: Represents a Frontmatter.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
package skill

type Frontmatter struct {
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
// Skill represents a complete Agent Skill.
//
// Summary: Represents a Skill.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Skill struct {
	// Frontmatter contains the metadata parsed from the YAML frontmatter.
	Frontmatter `yaml:",inline"`

	// Instructions is the Markdown content following the frontmatter.
	Instructions string `json:"instructions"`

	// Path is the absolute path to the skill directory on the filesystem.
	Path string `json:"path,omitempty"`

	// Assets is a list of relative paths to assets (scripts, references, etc.)
	// This is populated by scanning the directory.
	Assets []string `json:"assets,omitempty"`
}
