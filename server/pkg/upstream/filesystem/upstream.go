// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package filesystem provides the filesystem upstream implementation.
package filesystem

import (
	"context"
	"crypto/sha256"
	"bufio"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
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
//
//nolint:gocyclo
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

	// Create the filesystem backend
	fs, err := u.createFilesystem(fsService)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create filesystem: %w", err)
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
							"name":   map[string]interface{}{"type": "string"},
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

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					return nil, err
				}

				entries, err := afero.ReadDir(fs, resolvedPath)
				if err != nil {
					return nil, err
				}

				resultList := []interface{}{}
				for _, entry := range entries {
					resultList = append(resultList, map[string]interface{}{
						"name":   entry.Name(),
						"is_dir": entry.IsDir(),
						"size":   entry.Size(),
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

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					return nil, err
				}

				// Check if it's a directory
				info, err := fs.Stat(resolvedPath)
				if err != nil {
					return nil, err
				}
				if info.IsDir() {
					return nil, fmt.Errorf("path is a directory")
				}

				content, err := afero.ReadFile(fs, resolvedPath)
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

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					// Check if parent directory is allowed if file doesn't exist yet
					// resolvePath usually checks validity of prefix.
					return nil, err
				}

				// Ensure parent directory exists
				parentDir := filepath.Dir(resolvedPath)
				if err := fs.MkdirAll(parentDir, 0755); err != nil {
					return nil, fmt.Errorf("failed to create parent directory: %w", err)
				}

				if err := afero.WriteFile(fs, resolvedPath, []byte(content), 0600); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "delete_file",
			Description: "Delete a file or empty directory.",
			Input: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to delete."},
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

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					return nil, err
				}

				if err := fs.Remove(resolvedPath); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "search_files",
			Description: "Search for a text pattern in files within a directory.",
			Input: map[string]interface{}{
				"path":    map[string]interface{}{"type": "string", "description": "The root directory to search."},
				"pattern": map[string]interface{}{"type": "string", "description": "The regular expression to search for."},
			},
			Output: map[string]interface{}{
				"matches": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"file":         map[string]interface{}{"type": "string"},
							"line_number":  map[string]interface{}{"type": "integer"},
							"line_content": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}
				patternStr, ok := args["pattern"].(string)
				if !ok {
					return nil, fmt.Errorf("pattern is required")
				}

				re, err := regexp.Compile(patternStr)
				if err != nil {
					return nil, fmt.Errorf("invalid regex pattern: %w", err)
				}

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					return nil, err
				}

				matches := []map[string]interface{}{}
				maxMatches := 100
				matchCount := 0

				err = afero.Walk(fs, resolvedPath, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						// Skip unreadable files
						return nil
					}
					if matchCount >= maxMatches {
						return filepath.SkipDir
					}
					if info.IsDir() {
						// Skip hidden directories like .git
						if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".." {
							return filepath.SkipDir
						}
						return nil
					}
					// Skip large files (e.g., > 10MB)
					if info.Size() > 10*1024*1024 {
						return nil
					}

					// Read file
					f, err := fs.Open(filePath)
					if err != nil {
						return nil
					}
					defer func() { _ = f.Close() }()

					// Check for binary
					// Read first 512 bytes
					buffer := make([]byte, 512)
					n, _ := f.Read(buffer)
					if n > 0 {
						contentType := http.DetectContentType(buffer[:n])
						if contentType == "application/octet-stream" {
							return nil
						}
						// Reset seeker
						if _, err := f.Seek(0, 0); err != nil {
							return nil
						}
					}

					scanner := bufio.NewScanner(f)
					lineNum := 0
					for scanner.Scan() {
						lineNum++
						line := scanner.Text()
						if re.MatchString(line) {
							// Relativize path
							relPath, _ := filepath.Rel(resolvedPath, filePath)
							if relPath == "" {
								relPath = filepath.Base(filePath)
							}

							matches = append(matches, map[string]interface{}{
								"file":         relPath,
								"line_number":  lineNum,
								"line_content": strings.TrimSpace(line),
							})
							matchCount++
							if matchCount >= maxMatches {
								return filepath.SkipDir
							}
						}
					}
					return nil
				})

				if err != nil && err != filepath.SkipDir {
					return nil, err
				}

				return map[string]interface{}{"matches": matches}, nil
			},
		},
		{
			Name:        "get_file_info",
			Description: "Get information about a file or directory.",
			Input: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path."},
			},
			Output: map[string]interface{}{
				"name":     map[string]interface{}{"type": "string"},
				"is_dir":   map[string]interface{}{"type": "boolean"},
				"size":     map[string]interface{}{"type": "integer"},
				"mod_time": map[string]interface{}{"type": "string"},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				path, ok := args["path"].(string)
				if !ok {
					return nil, fmt.Errorf("path is required")
				}

				resolvedPath, err := u.resolvePath(path, fsService)
				if err != nil {
					return nil, err
				}

				info, err := fs.Stat(resolvedPath)
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
			Description: "List the allowed root directories. (Deprecated with afero usage)",
			Input:       map[string]interface{}{},
			Output: map[string]interface{}{
				"roots": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
				// With afero, we might just have one root or multiple mounts.
				// For backward compatibility, we can list keys from RootPaths if available.
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
			Name:      proto.String(toolName),
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

func (u *Upstream) createFilesystem(config *configv1.FilesystemUpstreamService) (afero.Fs, error) {
	var baseFs afero.Fs

	// Determine the backend filesystem
	switch config.FilesystemType.(type) {
	case *configv1.FilesystemUpstreamService_Tmpfs:
		baseFs = afero.NewMemMapFs()

	case *configv1.FilesystemUpstreamService_Http:
		return nil, fmt.Errorf("http filesystem is not yet supported")

	case *configv1.FilesystemUpstreamService_Zip:
		// To support zipfs, we need the file path. But zipfs in afero is separate.
		// import "github.com/spf13/afero/zipfs"
		// zipfs.New(zipReader)
		// This requires opening the file first.
		return nil, fmt.Errorf("zip filesystem is not yet supported")

	case *configv1.FilesystemUpstreamService_Gcs:
		// Requires external package gcsfs
		return nil, fmt.Errorf("gcs filesystem is not yet supported")

	case *configv1.FilesystemUpstreamService_Sftp:
		// Requires external package sftpfs
		return nil, fmt.Errorf("sftp filesystem is not yet supported")

	case *configv1.FilesystemUpstreamService_Os:
		baseFs = afero.NewOsFs()

	default:
		// Fallback to OsFs for backward compatibility if root_paths is set?
		// Or defaulting to OsFs.
		baseFs = afero.NewOsFs()
	}

	// Wrap with ReadOnly if requested
	if config.GetReadOnly() {
		baseFs = afero.NewReadOnlyFs(baseFs)
	}

	return baseFs, nil
}

// resolvePath determines the actual path to access based on the filesystem type.
// For OsFs, it resolves virtual path to real OS path using root_paths.
// For others, it uses the virtual path directly (cleaned).
func (u *Upstream) resolvePath(virtualPath string, config *configv1.FilesystemUpstreamService) (string, error) {
	switch config.FilesystemType.(type) {
	case *configv1.FilesystemUpstreamService_Tmpfs:
		// For MemMapFs, just clean the path. It's virtual.
		return filepath.Clean(virtualPath), nil

	case *configv1.FilesystemUpstreamService_Os:
		return u.validatePath(virtualPath, config.RootPaths)

	default:
		// Default (legacy) uses OsFs and validatePath
		return u.validatePath(virtualPath, config.RootPaths)
	}
}

// validatePath checks if the given virtual path resolves to a safe local path
// within one of the allowed root paths.
// This is used ONLY when we are mapping virtual paths to local OS paths (Legacy/OsFs mode).
func (u *Upstream) validatePath(virtualPath string, rootPaths map[string]string) (string, error) {
	if len(rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
	}

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
		return "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Resolve symlinks for the target path too
	targetPathCanonical, err := filepath.EvalSymlinks(targetPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, we need to find the deepest existing ancestor
			// to ensure that no symlinks in the path point outside the root.
			currentPath := targetPathAbs
			var existingPath string
			var remainingPath string

			for {
				dir := filepath.Dir(currentPath)
				if dir == currentPath {
					// Reached root without finding existing path (should not happen if realRoot exists)
					return "", fmt.Errorf("failed to resolve path (root not found): %s", targetPathAbs)
				}

				// Check if dir exists
				if _, err := os.Stat(dir); err == nil {
					existingPath = dir
					var relErr error
					remainingPath, relErr = filepath.Rel(existingPath, targetPathAbs)
					if relErr != nil {
						return "", fmt.Errorf("failed to calculate relative path: %w", relErr)
					}
					break
				}
				currentPath = dir
			}

			// Resolve symlinks for the existing ancestor
			existingPathCanonical, err := filepath.EvalSymlinks(existingPath)
			if err != nil {
				return "", fmt.Errorf("failed to resolve ancestor path symlinks: %w", err)
			}

			// Construct the canonical path
			targetPathCanonical = filepath.Join(existingPathCanonical, remainingPath)

			// Note: We don't check if the "remainingPath" contains ".." because filepath.Rel and Join should handle it,
			// and we are constructing it from absolute paths.
			// However, since the intermediate directories don't exist, they can't be symlinks pointing elsewhere.
		} else {
			return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
		}
	}

	// 3. Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	return targetPathCanonical, nil
}
