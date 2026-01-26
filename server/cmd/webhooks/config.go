// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// HookConfig defines a single webhook handler configuration.
type HookConfig struct {
	Path    string `yaml:"path"`
	Handler string `yaml:"handler"` // e.g., "markdown", "truncate", "paginate"
}

// Config defines the structure of the webhook server configuration.
type Config struct {
	Hooks []HookConfig `yaml:"hooks"`
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
