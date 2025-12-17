// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/tool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func newDocsCmd(fs afero.Fs) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate Markdown documentation for registered tools",
		Long: `Load the configuration and generate a Markdown file listing all registered tools,
their descriptions, and input schemas.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fs == nil {
				fs = afero.NewOsFs()
			}
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, fs); err != nil {
				return err
			}

			configPaths := cfg.ConfigPaths()
			if len(configPaths) == 0 {
				return fmt.Errorf("no configuration file provided")
			}

			// Initialize application
			appInstance := app.NewApplication()

			// Load configuration and register tools
			// We use ReloadConfig as it contains the logic to load from store and register to ToolManager
			if err := appInstance.ReloadConfig(fs, configPaths); err != nil {
				return fmt.Errorf("failed to load configuration and register tools: %w", err)
			}

			tools := appInstance.ToolManager.ListTools()

			// Sort tools by name for deterministic output
			sort.Slice(tools, func(i, j int) bool {
				return tools[i].Tool().GetName() < tools[j].Tool().GetName()
			})

			markdown := generateMarkdown(tools)

			outputPath, _ := cmd.Flags().GetString("output")
			if outputPath != "" {
				if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Documentation generated at %s\n", outputPath)
			} else {
				fmt.Println(markdown)
			}

			return nil
		},
	}

	docsCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")

	return docsCmd
}

func generateMarkdown(tools []tool.Tool) string {
	var sb strings.Builder
	sb.WriteString("# Tool Documentation\n\n")
	sb.WriteString("This document lists the tools available in the MCP Any server configuration.\n\n")

	if len(tools) == 0 {
		sb.WriteString("*No tools registered.*\n")
		return sb.String()
	}

	for _, t := range tools {
		toolDef := t.Tool()
		sb.WriteString(fmt.Sprintf("## %s\n\n", toolDef.GetName()))

		if toolDef.GetDisplayName() != "" {
			sb.WriteString(fmt.Sprintf("**Display Name:** %s\n\n", toolDef.GetDisplayName()))
		}

		if toolDef.GetDescription() != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", toolDef.GetDescription()))
		}

		sb.WriteString("### Input Schema\n\n")
		inputSchema := toolDef.GetInputSchema()
		if inputSchema == nil && toolDef.GetAnnotations() != nil {
			inputSchema = toolDef.GetAnnotations().GetInputSchema()
		}

		if inputSchema != nil {
			// Convert input schema to JSON string
			// We use protojson to marshal the struct
			jsonBytes, err := protojson.Marshal(inputSchema)
			if err == nil {
				// Pretty print the JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, jsonBytes, "", "  "); err == nil {
					sb.WriteString("```json\n")
					sb.WriteString(prettyJSON.String())
					sb.WriteString("\n```\n\n")
				} else {
					sb.WriteString("```json\n")
					sb.WriteString(string(jsonBytes))
					sb.WriteString("\n```\n\n")
				}
			} else {
				sb.WriteString("_Error marshaling input schema_\n\n")
			}
		} else {
			sb.WriteString("_No input schema defined_\n\n")
		}
	}

	return sb.String()
}
