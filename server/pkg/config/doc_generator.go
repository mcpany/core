package config

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// GenerateDocumentation generates Markdown documentation for the tools defined in the configuration.
// It iterates through the configured upstream services, registers their tools, and produces a
// Markdown formatted string describing each tool and its input schema.
//
// Parameters:
//   ctx: The context for the operation.
//   cfg: The server configuration containing upstream service definitions.
//
// Returns:
//   A string containing the generated Markdown documentation.
//   An error if documentation generation fails (e.g., if a service cannot be initialized).
func GenerateDocumentation(ctx context.Context, cfg *configv1.McpAnyServerConfig) (string, error) {
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)

	for _, serviceConfig := range cfg.GetUpstreamServices() {
		if serviceConfig.GetDisable() {
			continue
		}
		upstream, err := upstreamFactory.NewUpstream(serviceConfig)
		if err != nil {
			return "", fmt.Errorf("failed to create upstream %s: %w", serviceConfig.GetName(), err)
		}
		if upstream == nil {
			continue
		}
		// Register to populate toolManager
		_, _, _, err = upstream.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to register service %s: %v\n", serviceConfig.GetName(), err)
		}
	}

	tools := toolManager.ListTools()
	// Sort tools by name for consistent output
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Tool().GetName() < tools[j].Tool().GetName()
	})

	var sb strings.Builder
	sb.WriteString("# Available Tools\n\n")

	if len(tools) == 0 {
		sb.WriteString("No tools found.\n")
		return sb.String(), nil
	}

	for _, t := range tools {
		v1Tool := t.Tool()
		name := v1Tool.GetName()
		if v1Tool.GetServiceId() != "" {
			name = v1Tool.GetServiceId() + "." + name
		}
		sb.WriteString(fmt.Sprintf("## `%s`\n\n", name))
		if v1Tool.GetDisplayName() != "" {
			sb.WriteString(fmt.Sprintf("**%s**\n\n", v1Tool.GetDisplayName()))
		}
		if v1Tool.GetDescription() != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", v1Tool.GetDescription()))
		}

		sb.WriteString("### Inputs\n\n")

		var inputSchema *structpb.Struct
		if v1Tool.GetInputSchema() != nil {
			inputSchema = v1Tool.GetInputSchema()
		} else if v1Tool.GetAnnotations() != nil && v1Tool.GetAnnotations().GetInputSchema() != nil {
			inputSchema = v1Tool.GetAnnotations().GetInputSchema()
		}

		if inputSchema != nil {
			// Marshal to JSON
			opts := protojson.MarshalOptions{
				Multiline: true,
				Indent:    "  ",
			}
			jsonBytes, err := opts.Marshal(inputSchema)
			if err == nil {
				sb.WriteString("```json\n")
				sb.WriteString(string(jsonBytes))
				sb.WriteString("\n```\n\n")
			}
		}

		sb.WriteString("---\n\n")
	}

	return sb.String(), nil
}
