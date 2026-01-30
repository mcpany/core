// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill

// Frontmatter represents the YAML frontmatter of a SKILL.md file.
type Frontmatter struct {
	// Name is the name of the skill.
	Name string `yaml:"name" json:"name"`
	// Description is a short description of the skill.
	Description string `yaml:"description" json:"description"`
	// License is the license of the skill.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Compatibility specifies the compatibility requirements.
	Compatibility string `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	// Metadata contains arbitrary key-value metadata.
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	// AllowedTools is a list of tools allowed for this skill.
	AllowedTools []string `yaml:"allowed-tools,omitempty" json:"allowedTools,omitempty"`
}

// Skill represents a complete Agent Skill.
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
