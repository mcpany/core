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
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SkillResource adapts a Skill (or its asset) to the Resource interface.
//
// It provides a way to expose skill documentation and associated assets (like images or text files)
// as MCP resources, making them accessible to clients.
//
// Summary: Represents a SkillResource.
type SkillResource struct {
	skill     *skill.Skill
	assetPath string // Relative path to asset. If empty, represents the main SKILL.md

	mu            sync.RWMutex
	cachedContent []byte
	lastModTime   time.Time
}

// Ensure SkillResource implements resource.Resource.
var _ resource.Resource = &SkillResource{}

// NewSkillResource creates a new skill resource.
//
// Summary: Creates a new skill resource.
//
// Parameters:
//   - s (*skill.Skill): The s.
//
// Returns:
//   - *SkillResource: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewSkillResource(s *skill.Skill) *SkillResource {
	return &SkillResource{
		skill: s,
	}
}

// NewSkillAssetResource creates a new skill asset resource.
//
// Summary: Creates a new skill asset resource.
//
// Parameters:
//   - s (*skill.Skill): The s.
//   - assetPath (string): The asset path.
//
// Returns:
//   - *SkillResource: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewSkillAssetResource(s *skill.Skill, assetPath string) *SkillResource {
	return &SkillResource{
		skill:     s,
		assetPath: assetPath,
	}
}

// URI uRI uri.
//
// Summary: URI uri.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SkillResource) URI() string {
	if r.assetPath == "" {
		return fmt.Sprintf("skills://%s/SKILL.md", r.skill.Name)
	}
	return fmt.Sprintf("skills://%s/%s", r.skill.Name, r.assetPath)
}

// Name name name.
//
// Summary: Name name.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SkillResource) Name() string {
	if r.assetPath == "" {
		return fmt.Sprintf("Skill: %s", r.skill.Name)
	}
	return fmt.Sprintf("Skill Asset: %s (%s)", r.assetPath, r.skill.Name)
}

// Service service service.
//
// Summary: Service service.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SkillResource) Service() string {
	return "skills"
}

// Resource resource resource.
//
// Summary: Resource resource.
//
// Parameters:
//   None.
//
// Returns:
//   - *mcp.Resource: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// resolvePath determines the absolute path to the resource file.
func (r *SkillResource) resolvePath() (string, error) {
	if r.assetPath == "" {
		// Read main SKILL.md
		return filepath.Join(r.skill.Path, skill.SkillFileName), nil
	}

	// Read asset
	// Resolve the skill path to its canonical absolute path
	skillPath, err := filepath.Abs(r.skill.Path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute skill path: %w", err)
	}
	skillPath, err = filepath.EvalSymlinks(skillPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks for skill path: %w", err)
	}

	path := filepath.Join(skillPath, r.assetPath)

	// Security check: Path Traversal
	if err = validation.IsSecurePath(r.assetPath); err != nil {
		return "", fmt.Errorf("invalid asset path: %w", err)
	}

	// Resolve Symlinks in target path
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("asset does not exist: %w", os.ErrNotExist)
		}
		logging.GetLogger().Error("Failed to resolve asset path", "path", path, "error", err)
		return "", fmt.Errorf("failed to resolve asset path")
	}

	realPath, err = filepath.Abs(realPath)
	if err != nil {
		logging.GetLogger().Error("Failed to get absolute path", "path", realPath, "error", err)
		return "", fmt.Errorf("failed to resolve asset path")
	}

	// Security check: Jailbreak
	if !strings.HasPrefix(realPath, skillPath+string(os.PathSeparator)) && realPath != skillPath {
		return "", fmt.Errorf("invalid path: points outside skill directory")
	}

	return realPath, nil
}

// Read read read.
//
// Summary: Read read.
//
// Parameters:
//   - _ (context.Context): Unused parameter.
//
// Returns:
//   - *mcp.ReadResourceResult: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (r *SkillResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	realPath, err := r.resolvePath()
	if err != nil {
		return nil, err
	}

	// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
	// Optimization: Check file modification time to serve from cache if possible.
	info, err := os.Stat(realPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill file does not exist: %w", os.ErrNotExist)
		}
		logging.GetLogger().Error("Failed to stat skill file", "error", err)
		return nil, fmt.Errorf("failed to read skill file")
	}

	r.mu.RLock()
	// Check cache validity
	if !r.lastModTime.IsZero() && info.ModTime().Equal(r.lastModTime) && r.cachedContent != nil {
		content := r.cachedContent
		r.mu.RUnlock()
		return r.createResult(content)
	}
	r.mu.RUnlock()

	// Cache Miss: Read file
	content, err := os.ReadFile(realPath)
	if err != nil {
		logging.GetLogger().Error("Failed to read skill file", "error", err)
		return nil, fmt.Errorf("failed to read skill file")
	}

	// Update Cache
	r.mu.Lock()
	r.cachedContent = content
	r.lastModTime = info.ModTime()
	r.mu.Unlock()

	return r.createResult(content)
}

func (r *SkillResource) createResult(content []byte) (*mcp.ReadResourceResult, error) {
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

// Subscribe subscribe subscribe.
//
// Summary: Subscribe subscribe.
//
// Parameters:
//   - _ (context.Context): Unused parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (r *SkillResource) Subscribe(_ context.Context) error {
	// No-op for now
	return nil
}

// RegisterSkillResources registerSkillResources register skill resources.
//
// Summary: RegisterSkillResources register skill resources.
//
// Parameters:
//   - rm (resource.ManagerInterface): The rm.
//   - sm (*skill.Manager): The sm.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
