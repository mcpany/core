
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	ReadFileToolName   = "readFile"
	WriteFileToolName  = "writeFile"
	DeleteFileToolName = "deleteFile"
	ListFilesToolName  = "listFiles"
)

type Client struct {
	basePath string
}

func NewClient(basePath string) *Client {
	return &Client{basePath: basePath}
}

func (c *Client) GetTools(ctx context.Context) ([]*v1.Tool, error) {
	return []*v1.Tool{
		c.getReadFileTool(),
		c.getWriteFileTool(),
		c.getDeleteFileTool(),
		c.getListFilesTool(),
	}, nil
}

func (c *Client) CallTool(ctx context.Context, toolName string, args *structpb.Struct) (*structpb.Struct, error) {
	switch toolName {
	case ReadFileToolName:
		return c.readFile(ctx, args)
	case WriteFileToolName:
		return c.writeFile(ctx, args)
	case DeleteFileToolName:
		return c.deleteFile(ctx, args)
	case ListFilesToolName:
		return c.listFiles(ctx, args)
	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func s(val string) *string {
	return &val
}

func (c *Client) getReadFileTool() *v1.Tool {
	return &v1.Tool{
		Name:        s(ReadFileToolName),
		Description: s("Reads the contents of a file."),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"path": {Kind: &structpb.Value_StringValue{StringValue: "The path to the file to read."}},
			},
		},
		OutputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"content": {Kind: &structpb.Value_StringValue{StringValue: "The contents of the file."}},
			},
		},
	}
}

func (c *Client) getWriteFileTool() *v1.Tool {
	return &v1.Tool{
		Name:        s(WriteFileToolName),
		Description: s("Writes content to a file."),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"path":    {Kind: &structpb.Value_StringValue{StringValue: "The path to the file to write to."}},
				"content": {Kind: &structpb.Value_StringValue{StringValue: "The content to write to the file."}},
			},
		},
		OutputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"success": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
			},
		},
	}
}

func (c *Client) getDeleteFileTool() *v1.Tool {
	return &v1.Tool{
		Name:        s(DeleteFileToolName),
		Description: s("Deletes a file."),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"path": {Kind: &structpb.Value_StringValue{StringValue: "The path to the file to delete."}},
			},
		},
		OutputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"success": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
			},
		},
	}
}

func (c *Client) getListFilesTool() *v1.Tool {
	return &v1.Tool{
		Name:        s(ListFilesToolName),
		Description: s("Lists the files in a directory."),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"path": {Kind: &structpb.Value_StringValue{StringValue: "The path to the directory to list."}},
			},
		},
		OutputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"files": {Kind: &structpb.Value_ListValue{}},
			},
		},
	}
}

func (c *Client) readFile(ctx context.Context, args *structpb.Struct) (*structpb.Struct, error) {
	path := args.Fields["path"].GetStringValue()
	fullPath := filepath.Join(c.basePath, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"content": {Kind: &structpb.Value_StringValue{StringValue: string(content)}},
		},
	}, nil
}

func (c *Client) writeFile(ctx context.Context, args *structpb.Struct) (*structpb.Struct, error) {
	path := args.Fields["path"].GetStringValue()
	content := args.Fields["content"].GetStringValue()
	fullPath := filepath.Join(c.basePath, path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return nil, err
	}
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
		},
	}, nil
}

func (c *Client) deleteFile(ctx context.Context, args *structpb.Struct) (*structpb.Struct, error) {
	path := args.Fields["path"].GetStringValue()
	fullPath := filepath.Join(c.basePath, path)
	err := os.Remove(fullPath)
	if err != nil {
		return nil, err
	}
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
		},
	}, nil
}

func (c *Client) listFiles(ctx context.Context, args *structpb.Struct) (*structpb.Struct, error) {
	path := args.Fields["path"].GetStringValue()
	fullPath := filepath.Join(c.basePath, path)
	files, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	fileList := make([]*structpb.Value, 0, len(files))
	for _, file := range files {
		fileList = append(fileList, &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: file.Name()}})
	}
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"files": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: fileList}}},
		},
	}, nil
}
func (c *Client) GetUpstreamConfig() any {
	return nil
}

func (c *Client) GetPrompts(ctx context.Context) ([]*v1.Prompt, error) {
	return nil, nil
}
