// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ResolveRelativePaths resolves relative paths in the configuration relative to the base directory.
// This ensures that paths specified in a configuration file are interpreted relative to that file,
// rather than the current working directory of the server process.
func ResolveRelativePaths(cfg *configv1.McpAnyServerConfig, baseDir string) {
	if baseDir == "" {
		return
	}
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		// If we can't get absolute path, fallback to using baseDir as is (best effort)
		absBaseDir = baseDir
	}

	// Resolve Profile Secrets in Global Settings
	if cfg.GlobalSettings != nil {
		for _, profile := range cfg.GlobalSettings.ProfileDefinitions {
			resolveSecretMap(profile.Secrets, absBaseDir)
		}
	}

	// Resolve Upstream Services
	for _, service := range cfg.UpstreamServices {
		resolveUpstreamService(service, absBaseDir)
	}
}

func resolveUpstreamService(service *configv1.UpstreamServiceConfig, baseDir string) {
	if mcp := service.GetMcpService(); mcp != nil {
		if stdio := mcp.GetStdioConnection(); stdio != nil {
			resolveStdioConnection(stdio, baseDir)
		}
		if bundle := mcp.GetBundleConnection(); bundle != nil {
			resolveBundleConnection(bundle, baseDir)
		}
	}
	if cmd := service.GetCommandLineService(); cmd != nil {
		resolveCommandLineService(cmd, baseDir)
	}
}

func resolveStdioConnection(stdio *configv1.McpStdioConnection, baseDir string) {
	// Resolve Working Directory
	wd := stdio.GetWorkingDirectory()
	resolvedWD := resolveWorkingDirectory(wd, baseDir)
	stdio.WorkingDirectory = proto.String(resolvedWD)

	// Resolve Secrets in Env
	resolveSecretMap(stdio.Env, baseDir)
}

func resolveBundleConnection(bundle *configv1.McpBundleConnection, baseDir string) {
	// Resolve Bundle Path
	if bundle.GetBundlePath() != "" {
		bundle.BundlePath = proto.String(resolvePath(bundle.GetBundlePath(), baseDir))
	}

	// Resolve Secrets in Env
	resolveSecretMap(bundle.Env, baseDir)
}

func resolveCommandLineService(cmd *configv1.CommandLineUpstreamService, baseDir string) {
	// Resolve Working Directory
	wd := cmd.GetWorkingDirectory()
	resolvedWD := resolveWorkingDirectory(wd, baseDir)
	cmd.WorkingDirectory = proto.String(resolvedWD)
}

func resolveSecretMap(secrets map[string]*configv1.SecretValue, baseDir string) {
	for _, secret := range secrets {
		// Only resolve if it is a FilePath
		if secret.GetFilePath() != "" {
			secret.Value = &configv1.SecretValue_FilePath{
				FilePath: resolvePath(secret.GetFilePath(), baseDir),
			}
		}
	}
}

// resolvePath resolves a path relative to baseDir.
// If path is already absolute, it is returned as is.
// If path is relative, it is joined with baseDir.
func resolvePath(path string, baseDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// resolveWorkingDirectory resolves the working directory.
// If wd is empty, it returns baseDir (defaulting to the config file's directory).
// If wd is relative, it resolves it relative to baseDir.
// If wd is absolute, it returns it as is.
func resolveWorkingDirectory(wd string, baseDir string) string {
	if wd == "" {
		return baseDir
	}
	return resolvePath(wd, baseDir)
}
