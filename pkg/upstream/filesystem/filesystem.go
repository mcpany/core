package filesystem

import (
	"context"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/service/filesystem"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

type FileSystemUpstream struct {
	// Add any necessary fields here.
}

func NewFileSystemUpstream() upstream.Upstream {
	return &FileSystemUpstream{}
}

func (u *FileSystemUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	service := filesystem.NewFileSystemService(serviceConfig.GetFilesystemService())
	tools, err := service.GetTools()
	if err != nil {
		return "", nil, nil, err
	}

	for _, t := range tools {
		if err := toolManager.AddTool(t); err != nil {
			return "", nil, nil, err
		}
	}

	return serviceConfig.GetName(), nil, nil, nil
}
