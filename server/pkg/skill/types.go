// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill

// Frontmatter defines the core structure for frontmatter within the system.
//
// Summary: Frontmatter defines the core structure for frontmatter within the system.
//
// Fields:
//   - Contains the configuration and state properties required for Frontmatter functionality.
type Frontmatter struct {
	Name         string            `yaml:"name" json:"name"`
	Description  string            `yaml:"description" json:"description"`
	License      string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string           `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata     map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	AllowedTools []string          `yaml:"allowed-tools,omitempty" json:"allowedTools,omitempty"`
}

// Skill defines the core structure for skill within the system.
//
// Summary: Skill defines the core structure for skill within the system.
//
// Fields:
//   - Contains the configuration and state properties required for Skill functionality.
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
