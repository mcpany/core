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
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SkillResource adapts a Skill (or its asset) to the Resource interface.
type SkillResource struct {
	skill     *skill.Skill
	assetPath string // Relative path to asset. If empty, represents the main SKILL.md

	// ⚡ BOLT: Caching resolved path to avoid repeated I/O (evalSymlinks) on every read.
	cachedPath  string
	resolveOnce sync.Once
	resolveErr  error
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
		// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
		// Use cached resolution to avoid expensive filesystem calls.
		r.resolveOnce.Do(func() {
			// Read asset
			// Resolve the skill path to its canonical absolute path to prevent traversal via symlinks in the base path.
			var skillPath string
			skillPath, err = filepath.Abs(r.skill.Path)
			if err != nil {
				r.resolveErr = fmt.Errorf("failed to resolve absolute skill path: %w", err)
				return
			}
			skillPath, err = filepath.EvalSymlinks(skillPath)
			if err != nil {
				r.resolveErr = fmt.Errorf("failed to resolve symlinks for skill path: %w", err)
				return
			}

			path := filepath.Join(skillPath, r.assetPath)

			// Use centralized validation to ensure path is safe and within allowable bounds (which includes checking traversal)
			// However, validation.IsAllowedPath checks against CWD or AllowedPaths.
			// Here we specifically want to check if it is inside skillPath.
			// We can reuse validation.IsSecurePath to check for '..' traversal in the path string itself first.
			if err = validation.IsSecurePath(r.assetPath); err != nil {
				r.resolveErr = fmt.Errorf("invalid asset path: %w", err)
				return
			}

			// Now verify it resolves to inside skillPath.
			// We must resolve symlinks in the final path to ensure we don't traverse out via a symlink in the asset path.
			var realPath string
			realPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				if os.IsNotExist(err) {
					// Don't leak the full path in the error, but preserve error type
					r.resolveErr = fmt.Errorf("asset does not exist: %w", os.ErrNotExist)
					return
				}
				logging.GetLogger().Error("Failed to resolve asset path", "path", path, "error", err)
				r.resolveErr = fmt.Errorf("failed to resolve asset path")
				return
			}

			realPath, err = filepath.Abs(realPath)
			if err != nil {
				logging.GetLogger().Error("Failed to get absolute path", "path", realPath, "error", err)
				r.resolveErr = fmt.Errorf("failed to resolve asset path")
				return
			}

			// Check if the resolved path is inside the resolved skill path
			if !strings.HasPrefix(realPath, skillPath+string(os.PathSeparator)) && realPath != skillPath {
				r.resolveErr = fmt.Errorf("invalid path: points outside skill directory")
				return
			}

			r.cachedPath = realPath
		})

		if r.resolveErr != nil {
			return nil, r.resolveErr
		}

		content, err = os.ReadFile(r.cachedPath)
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
