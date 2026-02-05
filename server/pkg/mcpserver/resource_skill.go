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

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SkillResource adapts a Skill (or its asset) to the Resource interface.
//
// Summary: Represents a Skill file (SKILL.md or assets) as an MCP Resource.
type SkillResource struct {
	skill     *skill.Skill
	assetPath string // Relative path to asset. If empty, represents the main SKILL.md
}

// Ensure SkillResource implements resource.Resource.
var _ resource.Resource = &SkillResource{}

// NewSkillResource creates a new resource for the main SKILL.md.
//
// Summary: Initializes a new SkillResource pointing to the primary SKILL.md file of the skill.
//
// Parameters:
//   - s: *skill.Skill. The skill definition to create a resource for.
//
// Returns:
//   - *SkillResource: The initialized resource.
func NewSkillResource(s *skill.Skill) *SkillResource {
	return &SkillResource{
		skill: s,
	}
}

// NewSkillAssetResource creates a new resource for a skill asset.
//
// Summary: Initializes a new SkillResource pointing to an asset file within the skill.
//
// Parameters:
//   - s: *skill.Skill. The skill definition containing the asset.
//   - assetPath: string. The relative path to the asset file within the skill directory.
//
// Returns:
//   - *SkillResource: The initialized resource.
func NewSkillAssetResource(s *skill.Skill, assetPath string) *SkillResource {
	return &SkillResource{
		skill:     s,
		assetPath: assetPath,
	}
}

// URI returns the URI of the resource.
//
// Summary: Generates the unique URI for the skill resource.
// Format: skills://<skill_name>/SKILL.md or skills://<skill_name>/<asset_path>
//
// Returns:
//   - string: The resource URI.
func (r *SkillResource) URI() string {
	if r.assetPath == "" {
		return fmt.Sprintf("skills://%s/SKILL.md", r.skill.Name)
	}
	return fmt.Sprintf("skills://%s/%s", r.skill.Name, r.assetPath)
}

// Name returns the name of the resource.
//
// Summary: Generates a human-readable name for the resource.
//
// Returns:
//   - string: The name of the resource.
func (r *SkillResource) Name() string {
	if r.assetPath == "" {
		return fmt.Sprintf("Skill: %s", r.skill.Name)
	}
	return fmt.Sprintf("Skill Asset: %s (%s)", r.assetPath, r.skill.Name)
}

// Service returns the service associated with the resource.
//
// Summary: Identifies the service that owns this resource ("skills").
//
// Returns:
//   - string: The service identifier.
func (r *SkillResource) Service() string {
	return "skills"
}

// Resource returns the underlying MCP resource definition.
//
// Summary: Converts the SkillResource to an MCP Resource definition.
// Calculates MIME type based on file extension.
//
// Returns:
//   - *mcp.Resource: The MCP resource definition.
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
// Summary: Reads the file content from disk (SKILL.md or asset).
// Performs security checks to prevent directory traversal.
//
// Parameters:
//   - _ : context.Context. Unused.
//
// Returns:
//   - *mcp.ReadResourceResult: The resource content (text or blob).
//   - error: An error if the file cannot be read or security checks fail.
//
// Throws/Errors:
//   - Returns error if file path resolution fails or points outside the skill directory.
//   - Returns error if file does not exist or is unreadable.
func (r *SkillResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	var content []byte
	var err error

	if r.assetPath == "" {
		// Read main SKILL.md
		path := filepath.Join(r.skill.Path, skill.SkillFileName)
		content, err = os.ReadFile(path) //nolint:gosec
	} else {
		// Read asset
		// Resolve the skill path to its canonical absolute path to prevent traversal via symlinks in the base path.
		var skillPath string
		skillPath, err = filepath.Abs(r.skill.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute skill path: %w", err)
		}
		skillPath, err = filepath.EvalSymlinks(skillPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlinks for skill path: %w", err)
		}

		path := filepath.Join(skillPath, r.assetPath)

		// Use centralized validation to ensure path is safe and within allowable bounds (which includes checking traversal)
		// However, validation.IsAllowedPath checks against CWD or AllowedPaths.
		// Here we specifically want to check if it is inside skillPath.
		// We can reuse validation.IsSecurePath to check for '..' traversal in the path string itself first.
		if err = validation.IsSecurePath(r.assetPath); err != nil {
			return nil, fmt.Errorf("invalid asset path: %w", err)
		}

		// Now verify it resolves to inside skillPath.
		// We must resolve symlinks in the final path to ensure we don't traverse out via a symlink in the asset path.
		var realPath string
		realPath, err = filepath.EvalSymlinks(path)
		if err != nil {
			if os.IsNotExist(err) {
				// Don't leak the full path in the error, but preserve error type
				return nil, fmt.Errorf("asset does not exist: %w", os.ErrNotExist)
			}
			logging.GetLogger().Error("Failed to resolve asset path", "path", path, "error", err)
			return nil, fmt.Errorf("failed to resolve asset path")
		}

		realPath, err = filepath.Abs(realPath)
		if err != nil {
			logging.GetLogger().Error("Failed to get absolute path", "path", realPath, "error", err)
			return nil, fmt.Errorf("failed to resolve asset path")
		}

		// Check if the resolved path is inside the resolved skill path
		if !strings.HasPrefix(realPath, skillPath+string(os.PathSeparator)) && realPath != skillPath {
			return nil, fmt.Errorf("invalid path: points outside skill directory")
		}

		content, err = os.ReadFile(realPath) //nolint:gosec // Path is sanitized and verified to be within skill directory
	}

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill file does not exist: %w", os.ErrNotExist)
		}
		logging.GetLogger().Error("Failed to read skill file", "error", err)
		return nil, fmt.Errorf("failed to read skill file")
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
// Summary: No-op implementation for subscribing to skill resource changes.
//
// Parameters:
//   - _ : context.Context. Unused.
//
// Returns:
//   - error: Always returns nil.
func (r *SkillResource) Subscribe(_ context.Context) error {
	// No-op for now
	return nil
}

// RegisterSkillResources registers all skills from the manager into the resource manager.
//
// Summary: Iterates over all loaded skills and registers them (and their assets) as resources.
//
// Parameters:
//   - rm: resource.ManagerInterface. The resource manager to register resources with.
//   - sm: *skill.Manager. The skill manager to retrieve skills from.
//
// Returns:
//   - error: An error if listing skills fails.
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
