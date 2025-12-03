
package filesystem

import (
	"context"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/proto/config/v1"
)

type FilesystemUpstream struct{}

func NewFilesystemUpstream() *FilesystemUpstream {
	return &FilesystemUpstream{}
}

func (u *FilesystemUpstream) Register(
	ctx context.Context,
	serviceConfig *v1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*v1.ToolDefinition, []*v1.ResourceDefinition, error) {
	if err := Validate(serviceConfig.GetFilesystemService()); err != nil {
		return "", nil, nil, err
	}
	client := NewClient(serviceConfig.GetFilesystemService().GetBasePath())
	tools, err := client.GetTools(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	toolDefs := make([]*v1.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		toolDefs = append(toolDefs, &v1.ToolDefinition{
			Name: t.Name,
		})
	}
	return serviceConfig.GetName(), toolDefs, nil, nil
}
