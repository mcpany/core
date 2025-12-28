// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	"github.com/spf13/afero"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	log := logging.GetLogger()

	// Verify file exists
	absPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Configuration file not found: %s\n", absPath)
		os.Exit(1)
	}

	log.Info("Validating configuration", "path", absPath)

	// Use Config FileStore
	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, []string{absPath})

	// We validate as "server" type as that's the main use case for config validation
	_, err = config.LoadServices(store, "server")
	if err != nil {
		fmt.Printf("❌ Configuration validation failed:\n%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Configuration is valid.")
}
