// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package diagnostic provides tools for diagnosing server configuration and environment issues before startup.
package diagnostic

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
)

// CheckPortAvailability checks if the given address is available for listening.
func CheckPortAvailability(ctx context.Context, address string) error {
	// If address doesn't contain port, we can't really check, but assume it's formatted as host:port
	if !strings.Contains(address, ":") {
		return nil // Invalid format, let server handle it or default
	}

	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("port binding check failed for %s: %w", address, err)
	}
	_ = l.Close()
	return nil
}

// EnhanceConfigError tries to make configuration errors more human-readable.
func EnhanceConfigError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()

	// Handle proto unknown field errors
	if strings.Contains(msg, "unknown field") {
		// proto: (line 1:2): unknown field "services"
		parts := strings.Split(msg, "unknown field")
		if len(parts) > 1 {
			field := strings.Trim(parts[1], " \"")
			// The field name might have trailing characters like ")" or newline
			field = strings.TrimRight(field, ") \n")
			// Re-trim quotes if they were inside parens
			field = strings.Trim(field, "\"")
			suggestion := suggestField(field)
			return fmt.Errorf("configuration error: unknown field '%s'. %s", field, suggestion)
		}
	}

	// Handle YAML/JSON syntax errors
	if strings.Contains(msg, "yaml: line") || strings.Contains(msg, "json: syntax error") {
		return fmt.Errorf("configuration syntax error: %w", err)
	}

	// Handle wrapped errors
	if wrapped := strings.Split(msg, ": "); len(wrapped) > 1 {
		// Sometimes errors are like "failed to load config ... : proto: ..."
		// We want to extract the inner meaning.
		for _, part := range wrapped {
			if strings.Contains(part, "unknown field") {
				return EnhanceConfigError(fmt.Errorf("%s", part))
			}
		}
	}

	return err
}

func suggestField(field string) string {
	// Simple mapping for common mistakes
	commonMistakes := map[string]string{
		"services":         "Did you mean 'upstream_services'?",
		"mcpListenAddress": "Did you mean 'global_settings.mcp_listen_address'?",
		"port":             "Did you mean 'global_settings.mcp_listen_address' or 'grpc_port'?",
	}
	if suggestion, ok := commonMistakes[field]; ok {
		return suggestion
	}
	return "Please check the configuration reference."
}

// RunDiagnostics runs a suite of startup checks.
func RunDiagnostics(ctx context.Context, fs afero.Fs, configPaths []string, bindAddress string) error {
	// 1. Check Config Validity
	store := config.NewFileStore(fs, configPaths)
	// We just want to check if it parses correctly.
	_, err := config.LoadServices(ctx, store, "server")
	if err != nil {
		return EnhanceConfigError(err)
	}

	// 2. Check Port Availability
	if err := CheckPortAvailability(ctx, bindAddress); err != nil {
		return fmt.Errorf("network check failed: %w\n\tIs another instance of MCP Any running?", err)
	}

	return nil
}
