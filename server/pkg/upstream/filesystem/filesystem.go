// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package filesystem provides a filesystem upstream service.
package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	opList  = "list"
	opRead  = "read"
	opWrite = "write"
	opInfo  = "info"
)

// Upstream implements the upstream.Upstream interface for filesystem.
type Upstream struct{}

// NewUpstream creates a new filesystem upstream.
func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

// Shutdown shuts down the upstream.
func (u *Upstream) Shutdown(_ context.Context) error {
	return nil
}

// Register registers the filesystem tools.
func (u *Upstream) Register(
	_ context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName

	fsConfig := serviceConfig.GetFilesystemService()
	if fsConfig == nil {
		return "", nil, nil, fmt.Errorf("filesystem service config is nil")
	}

	rootPath := fsConfig.GetRootPath()
	if rootPath == "" {
		return "", nil, nil, fmt.Errorf("filesystem root_path is required")
	}
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to resolve absolute path for root_path: %w", err)
	}
	// Update config with absolute path
	fsConfig.RootPath = &absRoot

	// Check if directory exists
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to access root_path %q: %w", absRoot, err)
	}
	if !info.IsDir() {
		return "", nil, nil, fmt.Errorf("root_path %q is not a directory", absRoot)
	}

	infoStruct := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, infoStruct)

	// Register tools
	discoveredTools := []*configv1.ToolDefinition{}

	tools := []struct {
		name        string
		description string
		inputs      map[string]interface{}
		op          string
		write       bool
	}{
		{
			name:        "list_files",
			description: "List files and directories in the specified path",
			inputs: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to list, relative to root. Defaults to root."},
			},
			op: opList,
		},
		{
			name:        "read_file",
			description: "Read the content of a file",
			inputs: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to the file, relative to root"},
			},
			op: opRead,
		},
		{
			name:        "get_file_info",
			description: "Get metadata about a file or directory",
			inputs: map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "The path to the file or directory, relative to root"},
			},
			op: opInfo,
		},
		{
			name:        "write_file",
			description: "Write content to a file",
			inputs: map[string]interface{}{
				"path":    map[string]interface{}{"type": "string", "description": "The path to the file, relative to root"},
				"content": map[string]interface{}{"type": "string", "description": "The content to write"},
			},
			op:    opWrite,
			write: true,
		},
	}

	for _, t := range tools {
		if t.write && fsConfig.GetReadOnly() {
			continue
		}

		inputProps, _ := structpb.NewStruct(t.inputs)
		requiredList := []*structpb.Value{}
		if _, ok := t.inputs["path"]; ok {
			// path is optional for list (defaults to root), but required for others usually.
			// Let's make it optional for list, required for others.
			if t.op != opList {
				requiredList = append(requiredList, structpb.NewStringValue("path"))
			}
		}

		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(inputProps),
				"required":   structpb.NewListValue(&structpb.ListValue{Values: requiredList}),
			},
		}
		if t.op == opWrite {
			// Add content to required
			inputSchema.Fields["required"].GetListValue().Values = append(inputSchema.Fields["required"].GetListValue().Values, structpb.NewStringValue("content"))
		}

		outputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type": "object",
		})

		toolDef := &configv1.ToolDefinition{
			Name:        proto.String(t.name),
			Description: proto.String(t.description),
		}

		callable := &fsCallable{
			rootPath: absRoot,
			readOnly: fsConfig.GetReadOnly(),
			op:       t.op,
		}

		newTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to create tool %s: %w", t.name, err)
		}

		if err := toolManager.AddTool(newTool); err != nil {
			return "", nil, nil, fmt.Errorf("failed to add tool %s: %w", t.name, err)
		}
		discoveredTools = append(discoveredTools, toolDef)
	}

	log.Info(
		"Registered filesystem service",
		"serviceID",
		serviceID,
		"rootPath",
		absRoot,
		"toolsAdded",
		len(discoveredTools),
	)

	return serviceID, discoveredTools, nil, nil
}

type fsCallable struct {
	rootPath string
	readOnly bool
	op       string
}

func (c *fsCallable) Call(_ context.Context, req *tool.ExecutionRequest) (any, error) {
	var inputs map[string]interface{}
	if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
			return nil, fmt.Errorf("invalid inputs: %w", err)
		}
	}

	pathParam := "."
	if p, ok := inputs["path"].(string); ok {
		pathParam = p
	} else if c.op != opList {
		return nil, fmt.Errorf("path is required")
	}

	// Security check
	targetPath := filepath.Clean(filepath.Join(c.rootPath, pathParam))

	// Ensure targetPath starts with rootPath to prevent traversal
	rel, err := filepath.Rel(c.rootPath, targetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("access denied: path outside root")
	}

	switch c.op {
	case opList:
		return listFiles(targetPath)
	case opRead:
		return readFile(targetPath)
	case opWrite:
		if c.readOnly {
			return nil, fmt.Errorf("read-only mode")
		}
		content, ok := inputs["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content is required")
		}
		return writeFile(targetPath, content)
	case opInfo:
		return getFileInfo(targetPath)
	}
	return nil, fmt.Errorf("unknown operation")
}

func listFiles(path string) (any, error) {

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		result = append(result, name)
	}
	return result, nil
}

func readFile(path string) (any, error) {
	//nolint:gosec // Path is validated to be within root_path
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}

func writeFile(path string, content string) (any, error) {
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		return nil, err
	}
	return "success", nil
}

func getFileInfo(path string) (any, error) {

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"name":     info.Name(),
		"size":     info.Size(),
		"mode":     info.Mode().String(),
		"is_dir":   info.IsDir(),
		"mod_time": info.ModTime().String(),
	}, nil
}
