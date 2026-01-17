// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ResolveRelativePaths traverses the configuration and resolves relative file paths
// to absolute paths based on the provided base directory.
// This ensures that paths specified in configuration files are relative to the file itself,
// regardless of the server's working directory.
func ResolveRelativePaths(config *configv1.McpAnyServerConfig, baseDir string) {
	if config == nil || baseDir == "" {
		return
	}

	// Resolve Global Settings
	if gs := config.GetGlobalSettings(); gs != nil {
		if audit := gs.GetAudit(); audit != nil {
			if audit.GetStorageType() == configv1.AuditConfig_STORAGE_TYPE_FILE {
				audit.OutputPath = stringPtr(resolvePath(audit.GetOutputPath(), baseDir))
			}
		}
		if gc := gs.GetGcSettings(); gc != nil {
			for i, path := range gc.GetPaths() {
				gc.Paths[i] = resolvePath(path, baseDir)
			}
		}
		// GlobalSettings does not have MTLS directly.
	}

	// Resolve Upstream Services
	for _, service := range config.GetUpstreamServices() {
		resolveUpstreamService(service, baseDir)
	}
}

func resolveUpstreamService(service *configv1.UpstreamServiceConfig, baseDir string) {
	// Resolve Auth
	if auth := service.GetUpstreamAuth(); auth != nil {
		resolveAuthentication(auth, baseDir)
	}

	// Resolve specific service types
	if cmdService := service.GetCommandLineService(); cmdService != nil {
		cmdService.Command = stringPtr(resolveCommand(cmdService.GetCommand(), baseDir))
		if cmdService.GetWorkingDirectory() != "" {
			cmdService.WorkingDirectory = stringPtr(resolvePath(cmdService.GetWorkingDirectory(), baseDir))
		}
		// Resolve volume mounts in container environment - skipped for now
	} else if mcpService := service.GetMcpService(); mcpService != nil {
		if stdio := mcpService.GetStdioConnection(); stdio != nil {
			stdio.Command = stringPtr(resolveCommand(stdio.GetCommand(), baseDir))
			if stdio.GetWorkingDirectory() != "" {
				stdio.WorkingDirectory = stringPtr(resolvePath(stdio.GetWorkingDirectory(), baseDir))
			}
			resolveStdioArgs(stdio.GetCommand(), stdio.GetArgs(), baseDir)
			resolveSecretMap(stdio.GetEnv(), baseDir)
		} else if bundle := mcpService.GetBundleConnection(); bundle != nil {
			bundle.BundlePath = stringPtr(resolvePath(bundle.GetBundlePath(), baseDir))
			resolveSecretMap(bundle.GetEnv(), baseDir)
		} else if http := mcpService.GetHttpConnection(); http != nil {
			// No paths in HTTP connection usually
		}
	} else if sqlService := service.GetSqlService(); sqlService != nil {
		// Heuristic for SQLite
		if sqlService.GetDriver() == "sqlite" || sqlService.GetDriver() == "sqlite3" {
			// Only resolve if it looks like a path (contains separator or extension)
			// and isn't a memory DSN.
			dsn := sqlService.GetDsn()
			if dsn != ":memory:" && !strings.HasPrefix(dsn, "file:") {
				sqlService.Dsn = stringPtr(resolvePath(dsn, baseDir))
			}
		}
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		// If spec_url is not a http url, assume it's a file path
		if openapiService.GetSpecUrl() != "" && !isURL(openapiService.GetSpecUrl()) {
			openapiService.SpecSource = &configv1.OpenapiUpstreamService_SpecUrl{
				SpecUrl: resolvePath(openapiService.GetSpecUrl(), baseDir),
			}
		}
	}
}

func resolveAuthentication(auth *configv1.Authentication, baseDir string) {
	if mtls := auth.GetMtls(); mtls != nil {
		resolveMTLS(mtls, baseDir)
	}
	// API Key, Bearer, etc. might have SecretValue which can be file path
	switch auth.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		resolveSecretValue(auth.GetApiKey().GetValue(), baseDir)
	case configv1.Authentication_BearerToken_case:
		resolveSecretValue(auth.GetBearerToken().GetToken(), baseDir)
	case configv1.Authentication_BasicAuth_case:
		resolveSecretValue(auth.GetBasicAuth().GetPassword(), baseDir)
	case configv1.Authentication_Oauth2_case:
		resolveSecretValue(auth.GetOauth2().GetClientId(), baseDir)
		resolveSecretValue(auth.GetOauth2().GetClientSecret(), baseDir)
	}
}

func resolveMTLS(mtls *configv1.MTLSAuth, baseDir string) {
	mtls.ClientCertPath = stringPtr(resolvePath(mtls.GetClientCertPath(), baseDir))
	mtls.ClientKeyPath = stringPtr(resolvePath(mtls.GetClientKeyPath(), baseDir))
	if mtls.GetCaCertPath() != "" {
		mtls.CaCertPath = stringPtr(resolvePath(mtls.GetCaCertPath(), baseDir))
	}
}

func resolveSecretMap(secrets map[string]*configv1.SecretValue, baseDir string) {
	for _, s := range secrets {
		resolveSecretValue(s, baseDir)
	}
}

func resolveSecretValue(secret *configv1.SecretValue, baseDir string) {
	if secret == nil {
		return
	}
	if secret.WhichValue() == configv1.SecretValue_FilePath_case {
		secret.Value = &configv1.SecretValue_FilePath{
			FilePath: resolvePath(secret.GetFilePath(), baseDir),
		}
	}
}

// resolvePath resolves a path relative to baseDir.
// It assumes any path that is not absolute is relative.
func resolvePath(path string, baseDir string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// resolveCommand resolves a command path.
// It only resolves if the command contains a path separator, indicating it's a file path.
// Commands like "node", "python", "grep" are left as is to be resolved in PATH.
func resolveCommand(command string, baseDir string) string {
	if command == "" {
		return ""
	}
	// If it has a separator, it's a path (relative or absolute)
	if strings.Contains(command, string(filepath.Separator)) || strings.Contains(command, "/") {
		return resolvePath(command, baseDir)
	}
	return command
}

func resolveStdioArgs(command string, args []string, baseDir string) {
	// Only attempt to resolve args if the command looks like an interpreter
	baseCmd := filepath.Base(command)
	isInterpreter := false
	interpreters := []string{"python", "python3", "node", "deno", "bun", "ruby", "perl", "bash", "sh", "zsh", "go"}
	for _, i := range interpreters {
		if baseCmd == i || strings.HasPrefix(baseCmd, i) {
			isInterpreter = true
			break
		}
	}

	if !isInterpreter {
		return
	}

	isPython := strings.HasPrefix(baseCmd, "python")

	// Find the first non-flag argument
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if isPython && arg == "-m" {
				return // Python module mode, next arg is module name
			}
			continue
		}

		// URLs are valid for Deno/Bun
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			return
		}

		// Heuristic: if it has an extension, treat it as a script file
		ext := filepath.Ext(arg)
		if ext != "" {
			args[i] = resolvePath(arg, baseDir)
			return // Only resolve the first script argument
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
