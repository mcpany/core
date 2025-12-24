// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package filesystem provides the filesystem upstream implementation.
package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the upstream.Upstream interface for filesystem services.
type Upstream struct{}

// NewUpstream creates a new instance of FilesystemUpstream.
func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

// Shutdown implements the upstream.Upstream interface.
func (u *Upstream) Shutdown(_ context.Context) error {
	return nil
}

// Register processes the configuration for a filesystem service.
func (u *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName

	fsService := serviceConfig.GetFilesystemService()
	if fsService == nil {
		return "", nil, nil, fmt.Errorf("filesystem service config is nil")
	}

	// Validate root paths
	if len(fsService.RootPaths) == 0 {
		return "", nil, nil, fmt.Errorf("no root paths defined for filesystem service")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	// Define built-in tools
	tools := []*struct {
		Name        string
		Description string
		Input       map[string]interface{}
		Output      map[string]interface{}
		Handler     func(args map[string]interface{}) (map[string]interface{}, error)
	}{
		{
			Name:        "list_directory",
			Description: "List files and directories in a given path.",
			Input: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to list."},
			},
			Output: map[string]interface{}{
				"entries": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name":  map[string]interface{}{"type": "string"},
							"is_dir": map[string]interface{}{"type": "boolean"},
							"size":   map[string]interface{}{"type": "integer"},
						},
					},
				},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}
				realPath, err := u.validatePath(path, fsService.RootPaths)
				if err != nil {
					return nil, err
				}

				entries, err := os.ReadDir(realPath)
				if err != nil {
					return nil, err
				}

				resultList := []interface{}{}
				for _, entry := range entries {
					info, err := entry.Info()
					if err != nil {
						continue
					}
					resultList = append(resultList, map[string]interface{}{
						"name":   entry.Name(),
						"is_dir": entry.IsDir(),
						"size":   info.Size(),
					})
				}
				return map[string]interface{}{"entries": resultList}, nil
			},
		},
		{
			Name:        "read_file",
			Description: "Read the content of a file.",
			Input: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to the file."},
			},
			Output: map[string]interface{}{
				"content": map[string]interface{}{"type": "string", "description": "The file content."},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}
				realPath, err := u.validatePath(path, fsService.RootPaths)
				if err != nil {
					return nil, err
				}

				// Check if it's a directory
				info, err := os.Stat(realPath)
				if err != nil {
					return nil, err
				}
				if info.IsDir() {
					return nil, fmt.Errorf("path is a directory")
				}

				// #nosec G304 -- realPath is sanitized by validatePath which prevents traversal
				content, err := os.ReadFile(realPath)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"content": string(content)}, nil
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file.",
			Input: map[string]interface{}{
				"path":    map[string]interface{}{"type": "string", "description": "The path to the file."},
				"content": map[string]interface{}{"type": "string", "description": "The content to write."},
			},
			Output: map[string]interface{}{
				"success": map[string]interface{}{"type": "boolean"},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				if fsService.GetReadOnly() {
					return nil, fmt.Errorf("filesystem is read-only")
				}
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}
				content, ok := args["content"].(string)
				if !ok {
					return nil, fmt.Errorf("content is required")
				}

				realPath, err := u.validatePath(path, fsService.RootPaths)
				if err != nil {
					// Check if parent directory is allowed if file doesn't exist yet
					// But validatePath checks for prefix, so we need to construct what the real path *would* be
					// This is tricky. validatePath logic needs to handle non-existent files if we are writing.
					// Actually, validatePath implementation below handles checking if the resolved path starts with the real root.
					// It doesn't strictly require the file to exist, but filepath.Abs might resolve links.
					// Let's rely on validatePath to ensure safety.
					return nil, err
				}

				// #nosec G306 -- We explicitly want 0644 for files written by the tool
				if err := os.WriteFile(realPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "get_file_info",
			Description: "Get information about a file or directory.",
			Input: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path."},
			},
			Output: map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"is_dir":  map[string]interface{}{"type": "boolean"},
				"size":    map[string]interface{}{"type": "integer"},
				"mod_time": map[string]interface{}{"type": "string"},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}
				realPath, err := u.validatePath(path, fsService.RootPaths)
				if err != nil {
					return nil, err
				}
				info, err := os.Stat(realPath)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"name":     info.Name(),
					"is_dir":   info.IsDir(),
					"size":     info.Size(),
					"mod_time": info.ModTime().Format(time.RFC3339),
				}, nil
			},
		},
		{
			Name:        "list_allowed_directories",
			Description: "List the allowed root directories.",
			Input:       map[string]interface{}{},
			Output: map[string]interface{}{
				"roots": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				roots := []string{}
				for k := range fsService.RootPaths {
					roots = append(roots, k)
				}
				return map[string]interface{}{"roots": roots}, nil
			},
		},
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0)

	for _, t := range tools {
		toolName := t.Name

		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		if err != nil {
			log.Error("Failed to create input schema", "tool", toolName, "error", err)
			continue
		}

		outputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Output,
		})
		if err != nil {
			log.Error("Failed to create output schema", "tool", toolName, "error", err)
			continue
		}

		toolDef := configv1.ToolDefinition_builder{
			Name: proto.String(toolName),
			ServiceId: proto.String(serviceID),
		}.Build()

		handler := t.Handler
		callable := &fsCallable{handler: handler}

		// Create a callable tool
		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			log.Error("Failed to create callable tool", "tool", toolName, "error", err)
			continue
		}

		if err := toolManager.AddTool(callableTool); err != nil {
			log.Error("Failed to add tool", "tool", toolName, "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, toolDef)
	}

	log.Info("Registered filesystem service", "serviceID", serviceID, "tools", len(discoveredTools))
	return serviceID, discoveredTools, nil, nil
}

type fsCallable struct {
	handler func(args map[string]interface{}) (map[string]interface{}, error)
}

func (c *fsCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.handler(req.Arguments)
}

// validatePath checks if the given virtual path resolves to a safe local path
// within one of the allowed root paths.
func (u *Upstream) validatePath(virtualPath string, rootPaths map[string]string) (string, error) {
	// 1. Find the best matching root path (longest prefix match)
	var bestMatchVirtual string
	var bestMatchReal string

	for vRoot, rRoot := range rootPaths {
		// Ensure vRoot has a clean format
		cleanVRoot := vRoot
		if !strings.HasPrefix(cleanVRoot, "/") {
			cleanVRoot = "/" + cleanVRoot
		}

		// Ensure virtualPath starts with /
		checkPath := virtualPath
		if !strings.HasPrefix(checkPath, "/") {
			checkPath = "/" + checkPath
		}

		if strings.HasPrefix(checkPath, cleanVRoot) {
			if len(cleanVRoot) > len(bestMatchVirtual) {
				bestMatchVirtual = cleanVRoot
				bestMatchReal = rRoot
			}
		}
	}

	if bestMatchVirtual == "" {
		// Try fallback: if rootPaths has "/" key, use it.
		if val, ok := rootPaths["/"]; ok {
			bestMatchVirtual = "/"
			bestMatchReal = val
		} else {
			return "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}

	// 2. Resolve the path
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	// handle case where virtualPath matched exactly or with trailing slash
	relativePath = strings.TrimPrefix(relativePath, "/")

	realRootAbs, err := filepath.Abs(bestMatchReal)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Resolve symlinks in the root path to ensure we have the canonical path
	realRootCanonical, err := filepath.EvalSymlinks(realRootAbs)
	if err != nil {
		// If root doesn't exist, we can't really secure it, so error out.
		// However, for testing sometimes we use temp dirs.
		// If it fails, maybe fallback to Abs, but for security, EvalSymlinks is better.
		// If EvalSymlinks fails (e.g. doesn't exist), we should probably fail.
		return "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Resolve symlinks for the target path too, to ensure we don't jump out via a symlink *inside* the sandbox
	targetPathCanonical, err := filepath.EvalSymlinks(targetPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, we check the parent directory
			parent := filepath.Dir(targetPathAbs)
			parentCanonical, err := filepath.EvalSymlinks(parent)
			if err != nil {
				return "", fmt.Errorf("failed to resolve parent path symlinks: %w", err)
			}
			// Use the canonical parent + base name for check
			targetPathCanonical = filepath.Join(parentCanonical, filepath.Base(targetPathAbs))
		} else {
			return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
		}
	}

	// 3. Security Check: Ensure targetPathCanonical starts with realRootCanonical
	// Use explicit separator to prevent /root -> /root_evil
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	return targetPathCanonical, nil
}
