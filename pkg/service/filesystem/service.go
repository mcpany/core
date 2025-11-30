package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	pb "github.com/mcpany/core/proto/mcpany/service"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type FileSystemService struct {
	config *pb.FileSystemService
}

func NewFileSystemService(config *pb.FileSystemService) *FileSystemService {
	return &FileSystemService{config: config}
}

func (s *FileSystemService) Register(server *grpc.Server) {
	// Not a gRPC service, so nothing to do here.
}

func (s *FileSystemService) GetTools() ([]tool.Tool, error) {
	return []tool.Tool{
		tool.NewLocalTool(
			&v1.Tool{
				Name:        util.StringPtr("readFile"),
				Description: util.StringPtr("Reads the contents of a file."),
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"path": {Kind: &structpb.Value_StringValue{StringValue: "string"}},
					},
				},
			},
			s.readFile,
		),
		tool.NewLocalTool(
			&v1.Tool{
				Name:        util.StringPtr("writeFile"),
				Description: util.StringPtr("Writes content to a file."),
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"path":    {Kind: &structpb.Value_StringValue{StringValue: "string"}},
						"content": {Kind: &structpb.Value_StringValue{StringValue: "string"}},
					},
				},
			},
			s.writeFile,
		),
		tool.NewLocalTool(
			&v1.Tool{
				Name:        util.StringPtr("listFiles"),
				Description: util.StringPtr("Lists the files in a directory."),
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"path": {Kind: &structpb.Value_StringValue{StringValue: "string"}},
					},
				},
			},
			s.listFiles,
		),
	}, nil
}

func (s *FileSystemService) readFile(ctx context.Context, args *structpb.Struct) (*structpb.Value, error) {
	path, err := s.getSafePath(args.Fields["path"].GetStringValue())
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return structpb.NewStringValue(string(content)), nil
}

func (s *FileSystemService) writeFile(ctx context.Context, args *structpb.Struct) (*structpb.Value, error) {
	if s.config.ReadOnly {
		return nil, fmt.Errorf("file system is in read-only mode")
	}

	path, err := s.getSafePath(args.Fields["path"].GetStringValue())
	if err != nil {
		return nil, err
	}

	content := args.Fields["content"].GetStringValue()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return structpb.NewBoolValue(true), nil
}

func (s *FileSystemService) listFiles(ctx context.Context, args *structpb.Struct) (*structpb.Value, error) {
	path, err := s.getSafePath(args.Fields["path"].GetStringValue())
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var fileList []interface{}
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}

	listValue, err := structpb.NewList(fileList)
	if err != nil {
		return nil, err
	}

	return structpb.NewListValue(listValue), nil
}

func (s *FileSystemService) getSafePath(path string) (string, error) {
	root := s.config.RootDirectory
	if root == "" {
		root, _ = os.Getwd()
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(filepath.Join(absRoot, path))
	if err != nil {
		return "", err
	}

	if !filepath.HasPrefix(absPath, absRoot) {
		return "", fmt.Errorf("path is outside of the root directory")
	}

	return absPath, nil
}
