// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SkillResource adapts a Skill (or its asset) to the Resource interface.
type SkillResource struct {
	skill     *skill.Skill
	assetPath string // Relative path to asset. If empty, represents the main SKILL.md
}

// Ensure SkillResource implements resource.Resource.
var _ resource.Resource = &SkillResource{}

// NewSkillResource creates a new resource for the main SKILL.md.
//
// s is the s.
//
// Returns the result.
func NewSkillResource(s *skill.Skill) *SkillResource {
	return &SkillResource{
		skill: s,
	}
}

// NewSkillAssetResource creates a new resource for a skill asset.
//
// s is the s.
// assetPath is the assetPath.
//
// Returns the result.
func NewSkillAssetResource(s *skill.Skill, assetPath string) *SkillResource {
	return &SkillResource{
		skill:     s,
		assetPath: assetPath,
	}
}

// URI returns the URI of the resource.
//
// Returns the result.
func (r *SkillResource) URI() string {
	if r.assetPath == "" {
		return fmt.Sprintf("skills://%s/SKILL.md", r.skill.Name)
	}
	return fmt.Sprintf("skills://%s/%s", r.skill.Name, r.assetPath)
}

// Name returns the name of the resource.
//
// Returns the result.
func (r *SkillResource) Name() string {
	if r.assetPath == "" {
		return fmt.Sprintf("Skill: %s", r.skill.Name)
	}
	return fmt.Sprintf("Skill Asset: %s (%s)", r.assetPath, r.skill.Name)
}

// Service returns the service associated with the resource.
//
// Returns the result.
func (r *SkillResource) Service() string {
	return "skills"
}

// Resource returns the underlying MCP resource definition.
//
// Returns the result.
func (r *SkillResource) Resource() *mcp.Resource {
	mimeType := "text/markdown"
	if r.assetPath != "" {
		mimeType = mime.TypeByExtension(filepath.Ext(r.assetPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	return &mcp.Resource{
		Name:        r.Name(),
		URI:         r.URI(),
		MIMEType:    mimeType,
		Description: r.skill.Description,
	}
}

// Read returns the contents of the resource.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (r *SkillResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	var content []byte
	var err error

	if r.assetPath == "" {
		// Read main SKILL.md
		path := filepath.Join(r.skill.Path, skill.SkillFileName)
		content, err = os.ReadFile(path) //nolint:gosec
	} else {
		// Read asset
		// Sanitize and validate path to prevent traversal
		cleanAssetPath := filepath.Clean(r.assetPath)
		if strings.Contains(cleanAssetPath, "..") || strings.HasPrefix(cleanAssetPath, "/") || strings.HasPrefix(cleanAssetPath, "\\") {
			return nil, fmt.Errorf("invalid asset path: %s", r.assetPath)
		}

		skillPath := filepath.Clean(r.skill.Path)
		path := filepath.Join(skillPath, cleanAssetPath)
		// Double check resolved path to prevent traversal (e.g. /skill vs /skill-secret)
		if path != skillPath && !strings.HasPrefix(path, skillPath+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid path: points outside skill directory")
		}

		content, err = os.ReadFile(path) //nolint:gosec // Path is validated to be within skill directory
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	mimeType := r.Resource().MIMEType
	resourceContent := &mcp.ResourceContents{
		URI:      r.URI(),
		MIMEType: mimeType,
	}

	if isTextMime(mimeType) {
		resourceContent.Text = string(content)
	} else {
		resourceContent.Blob = content
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			resourceContent,
		},
	}, nil
}

func isTextMime(mimeType string) bool {
	baseMime, _, _ := strings.Cut(mimeType, ";")
	baseMime = strings.TrimSpace(baseMime)

	if strings.HasPrefix(baseMime, "text/") {
		return true
	}
	// Common text-based application types
	switch baseMime {
	case "application/json",
		"application/xml",
		"application/yaml",
		"application/x-yaml",
		"application/javascript",
		"application/ecmascript":
		return true
	}
	return false
}

// Subscribe subscribes to changes on the resource.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (r *SkillResource) Subscribe(_ context.Context) error {
	// No-op for now
	return nil
}

// RegisterSkillResources registers all skills from the manager into the resource manager.
//
// rm is the rm.
// sm is the sm.
//
// Returns an error if the operation fails.
func RegisterSkillResources(rm resource.ManagerInterface, sm *skill.Manager) error {
	skills, err := sm.ListSkills()
	if err != nil {
		return err
	}

	for _, s := range skills {
		// Register main SKILL.md
		rm.AddResource(NewSkillResource(s))

		// Register assets
		for _, asset := range s.Assets {
			rm.AddResource(NewSkillAssetResource(s, asset))
		}
	}
	return nil
}
