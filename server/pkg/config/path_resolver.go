// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ResolveRelativePaths traverses the configuration and resolves relative paths
// against the provided base directory (typically the directory of the config file).
// This ensures that paths like "./script.sh" or "certs/ca.pem" refer to files
// relative to the configuration file, rather than the server's working directory.
func ResolveRelativePaths(cfg *configv1.McpAnyServerConfig, baseDir string) {
	if cfg == nil || baseDir == "" {
		return
	}

	// Helper to resolve a single path (pointer)
	resolve := func(path *string) *string {
		if path == nil {
			return nil
		}
		return proto.String(resolvePath(baseDir, *path))
	}

	// Helper to resolve a single path (value)
	resolveStr := func(path string) string {
		return resolvePath(baseDir, path)
	}

	// Helper to resolve a command (pointer)
	resolveCmd := func(cmd *string) *string {
		if cmd == nil {
			return nil
		}
		return proto.String(resolveCommand(baseDir, *cmd))
	}

	for _, service := range cfg.UpstreamServices {
		// Common TLS Configs
		var tlsConfig *configv1.TLSConfig

		switch s := service.ServiceConfig.(type) {
		case *configv1.UpstreamServiceConfig_CommandLineService:
			svc := s.CommandLineService
			svc.Command = resolveCmd(svc.Command)
			svc.WorkingDirectory = resolve(svc.WorkingDirectory)

		case *configv1.UpstreamServiceConfig_McpService:
			svc := s.McpService
			switch conn := svc.ConnectionType.(type) {
			case *configv1.McpUpstreamService_StdioConnection:
				stdio := conn.StdioConnection
				stdio.Command = resolveCmd(stdio.Command)
				stdio.WorkingDirectory = resolve(stdio.WorkingDirectory)
			case *configv1.McpUpstreamService_BundleConnection:
				bundle := conn.BundleConnection
				bundle.BundlePath = resolve(bundle.BundlePath)
			case *configv1.McpUpstreamService_HttpConnection:
				tlsConfig = conn.HttpConnection.TlsConfig
			}

		case *configv1.UpstreamServiceConfig_GrpcService:
			svc := s.GrpcService
			tlsConfig = svc.TlsConfig
			for _, protoDef := range svc.ProtoDefinitions {
				switch r := protoDef.ProtoRef.(type) {
				case *configv1.ProtoDefinition_ProtoFile:
					if fPath, ok := r.ProtoFile.FileRef.(*configv1.ProtoFile_FilePath); ok {
						fPath.FilePath = resolveStr(fPath.FilePath)
					}
				case *configv1.ProtoDefinition_ProtoDescriptor:
					if fPath, ok := r.ProtoDescriptor.FileRef.(*configv1.ProtoDescriptor_FilePath); ok {
						fPath.FilePath = resolveStr(fPath.FilePath)
					}
				}
			}
			for _, coll := range svc.ProtoCollection {
				coll.RootPath = resolve(coll.RootPath)
			}

		case *configv1.UpstreamServiceConfig_HttpService:
			tlsConfig = s.HttpService.TlsConfig

		case *configv1.UpstreamServiceConfig_WebsocketService:
			tlsConfig = s.WebsocketService.TlsConfig

		case *configv1.UpstreamServiceConfig_WebrtcService:
			tlsConfig = s.WebrtcService.TlsConfig

		case *configv1.UpstreamServiceConfig_OpenapiService:
			svc := s.OpenapiService
			tlsConfig = svc.TlsConfig
			if spec, ok := svc.SpecSource.(*configv1.OpenapiUpstreamService_SpecUrl); ok {
				if isLocalFile(spec.SpecUrl) {
					spec.SpecUrl = resolveStr(spec.SpecUrl)
				}
			}

		case *configv1.UpstreamServiceConfig_FilesystemService:
			svc := s.FilesystemService
			for k, v := range svc.RootPaths {
				svc.RootPaths[k] = resolveStr(v)
			}
			switch fs := svc.FilesystemType.(type) {
			case *configv1.FilesystemUpstreamService_Zip:
				fs.Zip.FilePath = resolve(fs.Zip.FilePath)
			case *configv1.FilesystemUpstreamService_Sftp:
				fs.Sftp.KeyPath = resolve(fs.Sftp.KeyPath)
			}
		}

		if tlsConfig != nil {
			tlsConfig.CaCertPath = resolve(tlsConfig.CaCertPath)
			tlsConfig.ClientCertPath = resolve(tlsConfig.ClientCertPath)
			tlsConfig.ClientKeyPath = resolve(tlsConfig.ClientKeyPath)
		}
	}
}

func resolvePath(baseDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	// Check if it's a URL
	if strings.Contains(path, "://") {
		return path
	}
	return filepath.Join(baseDir, path)
}

func resolveCommand(baseDir, cmd string) string {
	if cmd == "" {
		return ""
	}
	// If it has a path separator, resolve it.
	// In Go, exec.Command searches PATH if no separator.
	if strings.ContainsRune(cmd, filepath.Separator) || strings.Contains(cmd, "/") {
		return resolvePath(baseDir, cmd)
	}
	return cmd
}

func isLocalFile(url string) bool {
	return !strings.Contains(url, "://") || strings.HasPrefix(url, "file://")
}
