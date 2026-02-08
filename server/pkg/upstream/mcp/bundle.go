// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Manifest represents the structure of the manifest.json file in an MCP bundle.
//
// Summary: represents the structure of the manifest.json file in an MCP bundle.
type Manifest struct {
	// ManifestVersion is the version of the manifest format.
	ManifestVersion string `json:"manifest_version"`
	// Name is the name of the bundle.
	Name string `json:"name"`
	// Version is the version of the bundle.
	Version string `json:"version"`
	// Description is a description of the bundle.
	Description string `json:"description"`
	// Server contains configuration for the MCP server within the bundle.
	Server ManifestServer `json:"server"`
	// UserConfig contains default configuration for the user.
	UserConfig json.RawMessage `json:"user_config"`
}

// ManifestServer represents the server configuration in the manifest.
//
// Summary: represents the server configuration in the manifest.
type ManifestServer struct {
	// Type is the type of the server (e.g., "node", "python").
	Type string `json:"type"`
	// EntryPoint is the entry point script or command for the server.
	EntryPoint string `json:"entry_point"`
	// McpConfig contains specific configuration for running the MCP server.
	McpConfig ManifestMcpConfig `json:"mcp_config"`
}

// ManifestMcpConfig represents the MCP configuration in the manifest.
//
// Summary: represents the MCP configuration in the manifest.
type ManifestMcpConfig struct {
	// Command is the command to run the server.
	Command string `json:"command"`
	// Args are the arguments to pass to the command.
	Args []string `json:"args"`
	// Env is a map of environment variables to set for the server.
	Env map[string]string `json:"env"`
}

// createAndRegisterMCPItemsFromBundle handles the registration of an MCP service
// from a bundle.
func (u *Upstream) createAndRegisterMCPItemsFromBundle(
	ctx context.Context,
	serviceID string,
	bundleConfig *configv1.McpBundleConnection,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ bool, // isReload
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	bundlePath := bundleConfig.GetBundlePath()
	if bundlePath == "" {
		return nil, nil, fmt.Errorf("bundle_path is required")
	}

	tempDir := filepath.Join(u.BundleBaseDir, serviceID)
	// Clean up old dir if exists
	_ = os.RemoveAll(tempDir)
	if err := unzipBundle(bundlePath, tempDir); err != nil {
		return nil, nil, fmt.Errorf("failed to unzip bundle: %w", err)
	}

	// 2. Read Manifest
	manifestPath := filepath.Join(tempDir, "manifest.json")
	manifestFile, err := os.Open(manifestPath) //nolint:gosec // Path constructed from secure temp dir
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open manifest.json: %w", err)
	}
	defer func() { _ = manifestFile.Close() }()

	var manifest Manifest
	if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to decode manifest.json: %w", err)
	}

	// 3. Determine Container Configuration
	containerMountPath := "/app/bundle"

	imageName := bundleConfig.GetContainerImage()
	if imageName == "" {
		imageName = inferImage(manifest.Server.Type)
		if imageName == "" {
			return nil, nil, fmt.Errorf("unable to infer container image for server type: %s", manifest.Server.Type)
		}
	}

	command := manifest.Server.McpConfig.Command
	args := manifest.Server.McpConfig.Args
	env := manifest.Server.McpConfig.Env
	if env == nil {
		env = make(map[string]string)
	}

	// Merge config env (overrides manifest env)
	// Merge config env (overrides manifest env)
	resolvedBundleEnv, err := util.ResolveSecretMap(ctx, bundleConfig.GetEnv(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve bundle secret env: %w", err)
	}

	for k, v := range resolvedBundleEnv {
		env[k] = v
	}

	// Default command/arg
	const (
		typeNode   = "node"
		typePython = "python"
	)

	switch manifest.Server.Type {
	case typeNode:
		// Node.js
		// We expect "mcp_config" -> command, args
		mcpConfig := manifest.Server.McpConfig
		command = mcpConfig.Command
		if command == "" {
			command = typeNode
		}
		args = mcpConfig.Args
		// If entry_point is specified, it might be an argument to node.
		if manifest.Server.EntryPoint != "" {
			// Prepend entry point to args if not present?
			// Usually entry point IS the script.
			args = append([]string{manifest.Server.EntryPoint}, args...)
		}
		imageName = "node:18-alpine"

	case typePython:
		command = typePython
		ep := manifest.Server.EntryPoint
		if ep == "" {
			ep = "main.py"
		}
		args = []string{filepath.Join(containerMountPath, ep)}
	case "uv":
		runCmd := []string{"uv", "run"}
		ep := manifest.Server.EntryPoint
		if ep != "" {
			runCmd = append(runCmd, filepath.Join(containerMountPath, ep))
		}
		command = runCmd[0]
		args = runCmd[1:]
	}
	// Variable substitution for ${__dirname}
	for i, arg := range args {
		args[i] = strings.ReplaceAll(arg, "${__dirname}", containerMountPath)
	}

	envList := make([]string, 0, len(env))
	for k, v := range env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	// 4. Construct Transport
	if stat, err := os.Stat(tempDir); err != nil {
		logging.GetLogger().Error("Bundle temp dir check failed", "path", tempDir, "error", err)
	} else {
		logging.GetLogger().Info("Bundle temp dir exists", "path", tempDir, "mode", stat.Mode())
	}
	transport := &BundleDockerTransport{
		Image:      imageName,
		Command:    command,
		Args:       args,
		Env:        envList,
		WorkingDir: containerMountPath,
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   tempDir,
				Target:   containerMountPath,
				ReadOnly: true,
			},
		},
	}

	// 5. Connect and Register
	mcpSdkClient, err := u.createMCPClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	var cs ClientSession
	if connectForTesting != nil {
		cs, err = connectForTesting(mcpSdkClient, ctx, transport, nil)
	} else {
		cs, err = mcpSdkClient.Connect(ctx, transport, nil)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MCP bundle service: %w", err)
	}
	defer func() { _ = cs.Close() }()

	// List Tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools from MCP bundle: %w", err)
	}

	bundleConn := &mcpConnection{
		client:          mcpSdkClient,
		bundleTransport: transport,
		sessionRegistry: u.sessionRegistry,
	}

	return u.processMCPItems(ctx, serviceID, listToolsResult, bundleConn, bundleConn, cs, toolManager, promptManager, resourceManager, serviceConfig)
}

func unzipBundle(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	if err := os.MkdirAll(destAbs, 0750); err != nil {
		return err
	}

	// Resolve symlinks in destination to ensure we have the canonical path
	destCanonical, err := filepath.EvalSymlinks(destAbs)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks for destination: %w", err)
	}

	var totalBytes int64
	const maxTotalSize = 5 * 1024 * 1024 * 1024 // 5GB

	for _, f := range r.File {
		// Join path and resolve absolute path
		fpath := filepath.Join(destCanonical, f.Name) //nolint:gosec // Path is validated immediately after

		// Check for Zip Slip (Zip Path Traversal)
		// We use HasPrefix on the clean, absolute paths.
		// Note: filepath.Join cleans the path.
		if !strings.HasPrefix(fpath, destCanonical+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0750); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0750); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()) //nolint:gosec // Path validated in loop
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		// Mitigate G110: Decompression bomb. Limit max file size to 1GB.
		const maxFileSize = 1 * 1024 * 1024 * 1024 // 1GB

		// Use io.LimitReader to prevent reading more than maxFileSize.
		// We read up to maxFileSize bytes.
		n, err := io.Copy(outFile, io.LimitReader(rc, maxFileSize))
		if err != nil {
			_ = outFile.Close()
			_ = rc.Close()
			return fmt.Errorf("failed to copy file content: %w", err)
		}

		totalBytes += n
		if totalBytes > maxTotalSize {
			_ = outFile.Close()
			_ = rc.Close()
			return fmt.Errorf("total decompressed size exceeds limit of %d bytes", maxTotalSize)
		}

		// Check if there is more data remaining (meaning it exceeded the limit)
		// We try to read 1 byte.
		buf := make([]byte, 1)
		read, _ := rc.Read(buf)
		if read > 0 || n == maxFileSize {
			// If n hit the limit, we should double check if we can read more.
			// But if n < maxFileSize, we are sure we consumed it all (given io.Copy finishes on EOF).
			// If n == maxFileSize, it might be EXACTLY maxFileSize or larger.
			if n == maxFileSize {
				// Try reading one more byte to confirm if it's larger
				if read > 0 {
					_ = outFile.Close()
					_ = rc.Close()
					return fmt.Errorf("file %s exceeds maximum allowed size of %d bytes", f.Name, maxFileSize)
				}
				// If read == 0, it means we hit EOF exactly at maxFileSize, which is fine.
			}
		}

		_ = outFile.Close()
		_ = rc.Close()
	}
	return nil
}

func inferImage(serverType string) string {
	switch serverType {
	case "node":
		return "node:18-alpine"
	case "python":
		return "python:3.11-slim"
	case "uv":
		return "ghcr.io/astral-sh/uv:python3.11-bookworm-slim"
	case "binary":
		return "debian:bookworm-slim"
	default:
		return ""
	}
}
