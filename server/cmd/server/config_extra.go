// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func runSnapshot(cmd *cobra.Command, _ []string) error {
	osFs := afero.NewOsFs()
	cfgSettings := config.GlobalSettings()
	if err := cfgSettings.Load(cmd, osFs); err != nil {
		return err
	}

	store := config.NewFileStore(osFs, cfgSettings.ConfigPaths())
	store.SetIgnoreMissingEnv(true)
	configs, err := config.LoadResolvedConfig(context.Background(), store)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	marshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}
	jsonBytes, err := marshaler.Marshal(configs)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func runFixEnv(cmd *cobra.Command, _ []string) error {
	osFs := afero.NewOsFs()
	cfgSettings := config.GlobalSettings()
	_ = cfgSettings.Load(cmd, osFs)

	paths := cfgSettings.ConfigPaths()
	if len(paths) == 0 {
		return fmt.Errorf("no config files found")
	}

	var missingVars []string
	seen := make(map[string]bool)
	// Match ${VAR_NAME}
	re := regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip if file not found (handled by loader usually, but here we scan best effort)
			continue
		}
		matches := re.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			varName := m[1]
			if !seen[varName] {
				if _, exists := os.LookupEnv(varName); !exists {
					missingVars = append(missingVars, varName)
				}
				seen[varName] = true
			}
		}
	}

	if len(missingVars) == 0 {
		fmt.Println("✅ All referenced environment variables are set.")
		return nil
	}

	fmt.Printf("Found %d missing environment variables:\n", len(missingVars))

	f, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer f.Close()

	reader := bufio.NewReader(os.Stdin)
	for _, v := range missingVars {
		fmt.Printf("Enter value for %s (leave empty to skip): ", v)
		val, _ := reader.ReadString('\n')
		val = strings.TrimSpace(val)
		if val != "" {
			if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", v, val)); err != nil {
				return err
			}
			fmt.Printf("Saved %s to .env\n", v)
		}
	}

	return nil
}
