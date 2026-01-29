package filesystem

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
)

func listDirectoryTool(prov provider.Provider, fs afero.Fs) filesystemToolDef {
	return filesystemToolDef{
		Name:        "list_directory",
		Description: "List files and directories in a given path.",
		Input: map[string]interface{}{
			"path": map[string]interface{}{"type": "string", "description": "The path to list."},
		},
		Output: map[string]interface{}{
			"entries": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":   map[string]interface{}{"type": "string"},
						"is_dir": map[string]interface{}{"type": "boolean"},
						"size":   map[string]interface{}{"type": "integer"},
					},
				},
			},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				return nil, err
			}

			entries, err := afero.ReadDir(fs, resolvedPath)
			if err != nil {
				return nil, err
			}

			resultList := []interface{}{}
			for _, entry := range entries {
				resultList = append(resultList, map[string]interface{}{
					"name":   entry.Name(),
					"is_dir": entry.IsDir(),
					"size":   entry.Size(),
				})
			}
			return map[string]interface{}{"entries": resultList}, nil
		},
	}
}

func getFileInfoTool(prov provider.Provider, fs afero.Fs) filesystemToolDef {
	return filesystemToolDef{
		Name:        "get_file_info",
		Description: "Get information about a file or directory.",
		Input: map[string]interface{}{
			"path": map[string]interface{}{"type": "string", "description": "The path."},
		},
		Output: map[string]interface{}{
			"name":     map[string]interface{}{"type": "string"},
			"is_dir":   map[string]interface{}{"type": "boolean"},
			"size":     map[string]interface{}{"type": "integer"},
			"mod_time": map[string]interface{}{"type": "string"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				return nil, err
			}

			info, err := fs.Stat(resolvedPath)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"name":     info.Name(),
				"is_dir":   info.IsDir(),
				"size":     info.Size(),
				"mod_time": info.ModTime().Format(time.RFC3339),
			}, nil
		},
	}
}
