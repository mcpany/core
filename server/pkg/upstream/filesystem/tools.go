package filesystem

import (
	"context"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
)

func getTools(prov provider.Provider, fs afero.Fs, readOnly bool, rootPaths map[string]string) []filesystemToolDef {
	return []filesystemToolDef{
		listDirectoryTool(prov, fs),
		readFileTool(prov, fs),
		writeFileTool(prov, fs, readOnly),
		moveFileTool(prov, fs, readOnly),
		deleteFileTool(prov, fs, readOnly),
		searchFilesTool(prov, fs),
		getFileInfoTool(prov, fs),
		{
			Name:        "list_allowed_directories",
			Description: "List the allowed root directories. (Deprecated with afero usage)",
			Input:       map[string]interface{}{},
			Output: map[string]interface{}{
				"roots": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			Handler: func(_ context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
				roots := []string{}
				for k := range rootPaths {
					roots = append(roots, k)
				}
				return map[string]interface{}{"roots": roots}, nil
			},
		},
	}
}
