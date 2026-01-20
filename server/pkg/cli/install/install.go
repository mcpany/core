// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package install provides the CLI command for installing MCP Any into AI clients.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Options holds the configuration for the install command.
type Options struct {
	Client      string
	ConfigPath  string
	LocalBinary bool
	Name        string
	Fs          afero.Fs
}

// MCPServer represents a single MCP server configuration.
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// NewCmdInstall returns the install command.
func NewCmdInstall() *cobra.Command {
	opts := &Options{
		Fs: afero.NewOsFs(),
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install MCP Any into your AI Client (e.g. Claude Desktop)",
		Long: `Automatically configures your AI Client to use MCP Any.
Currently supports Claude Desktop.

Examples:
  # Install for Claude Desktop using Docker (recommended)
  mcpany install --config-path ./my-config.yaml

  # Install for Claude Desktop using local binary
  mcpany install --local-binary --config-path ./my-config.yaml
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return Run(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Client, "client", "claude", "The client to configure (currently only 'claude' is supported)")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", "", "Path to your MCP Any configuration file (required)")
	cmd.Flags().BoolVar(&opts.LocalBinary, "local-binary", false, "Use the current binary path instead of Docker")
	cmd.Flags().StringVar(&opts.Name, "name", "mcpany", "The name of the server in the client configuration")

	_ = cmd.MarkFlagRequired("config-path")

	return cmd
}

// Run executes the install logic.
func Run(cmd *cobra.Command, opts *Options) error {
	// Validate inputs
	if opts.ConfigPath == "" {
		return fmt.Errorf("config-path is required")
	}

	// Resolve absolute path for config
	absConfigPath, err := filepath.Abs(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if config file exists
	exists, err := afero.Exists(opts.Fs, absConfigPath)
	if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}
	if !exists {
		return fmt.Errorf("configuration file not found at %s", absConfigPath)
	}

	// Handle Client Selection
	switch strings.ToLower(opts.Client) {
	case "claude":
		return installClaude(cmd, opts, absConfigPath)
	default:
		return fmt.Errorf("unsupported client: %s", opts.Client)
	}
}

func installClaude(cmd *cobra.Command, opts *Options, absConfigPath string) error {
	configPath, err := getClaudeConfigPath()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔍 Detected Claude Desktop configuration at: %s\n", configPath)

	// Ensure directory exists
	if err := opts.Fs.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Read existing config or create new
	// We use map[string]interface{} to preserve unknown fields
	config := make(map[string]interface{})
	exists, err := afero.Exists(opts.Fs, configPath)
	if err != nil {
		return fmt.Errorf("failed to check claude config: %w", err)
	}

	if exists {
		data, err := afero.ReadFile(opts.Fs, configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &config); err != nil {
				// If JSON is malformed, we might want to warn or backup.
				// For now, let's return error.
				return fmt.Errorf("failed to parse existing config file: %w", err)
			}
		}
	}

	// Extract or initialize mcpServers map
	var mcpServers map[string]interface{}
	if val, ok := config["mcpServers"]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			mcpServers = m
		} else {
			// mcpServers exists but is not a map? Reset it.
			mcpServers = make(map[string]interface{})
		}
	} else {
		mcpServers = make(map[string]interface{})
	}

	// Construct Server Config
	serverConfig, err := buildServerConfig(opts, absConfigPath)
	if err != nil {
		return err
	}

	mcpServers[opts.Name] = serverConfig
	config["mcpServers"] = mcpServers

	// Write back
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := afero.WriteFile(opts.Fs, configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✅ Successfully installed 'mcpany' to Claude Desktop configuration!\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🎉 You can now restart Claude Desktop to use your new tools.\n")

	return nil
}

func getClaudeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default: // Linux and others
		// Using standard XDG config location or ~/.config/Claude
		// Some users report ~/.config/Claude/claude_desktop_config.json
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
}

func buildServerConfig(opts *Options, absConfigPath string) (MCPServer, error) {
	if opts.LocalBinary {
		exe, err := os.Executable()
		if err != nil {
			return MCPServer{}, fmt.Errorf("failed to get current executable path: %w", err)
		}
		return MCPServer{
			Command: exe,
			Args: []string{
				"run",
				"--config-path", absConfigPath,
				"--stdio",
			},
		}, nil
	}

	// Docker Config
	// docker run -i --rm -v /absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml ghcr.io/mcpany/server:latest run --config-path /etc/mcpany/config.yaml --stdio

	configDir := filepath.Dir(absConfigPath)
	configFileName := filepath.Base(absConfigPath)
	containerMountPath := "/etc/mcpany"
	containerConfigPath := fmt.Sprintf("%s/%s", containerMountPath, configFileName)

	return MCPServer{
		Command: "docker",
		Args: []string{
			"run",
			"-i",
			"--rm",
			"-v", fmt.Sprintf("%s:%s", configDir, containerMountPath),
			"ghcr.io/mcpany/server:latest",
			"run",
			"--config-path", containerConfigPath,
			"--stdio",
		},
	}, nil
}
