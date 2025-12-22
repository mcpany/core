package protobufparser

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcpopt "github.com/mcpany/core/proto/mcp_options/v1"
)

// mockReflectionServer is a mock implementation of the gRPC reflection server.
type mockReflectionServer struct {
	reflectpb.UnimplementedServerReflectionServer
	streamReady chan struct{}
	stream      *mockReflectionServerStream
}

func (s *mockReflectionServer) ServerReflectionInfo(stream reflectpb.ServerReflection_ServerReflectionInfoServer) error {
	s.stream = &mockReflectionServerStream{stream}
	close(s.streamReady)
	<-stream.Context().Done()
	return nil
}

// mockReflectionServerStream is a mock implementation of the gRPC reflection server stream.
type mockReflectionServerStream struct {
	reflectpb.ServerReflection_ServerReflectionInfoServer
}

func TestParseProtoFromDefs_Extended(t *testing.T) {
	ctx := context.Background()

	t.Run("error on non-existent file path", func(t *testing.T) {
		protoDef := configv1.ProtoDefinition_builder{
			ProtoFile: configv1.ProtoFile_builder{
				FileName: proto.String("nonexistent.proto"),
				FilePath: proto.String("/path/to/nonexistent.proto"),
			}.Build(),
		}.Build()
		_, err := ParseProtoFromDefs(ctx, []*configv1.ProtoDefinition{protoDef}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read proto file from path")
	})

	t.Run("error on invalid regex in ProtoCollection", func(t *testing.T) {
		protoCollection := configv1.ProtoCollection_builder{
			RootPath:       proto.String("./"),
			PathMatchRegex: proto.String("["), // Invalid regex
		}.Build()
		_, err := ParseProtoFromDefs(ctx, nil, []*configv1.ProtoCollection{protoCollection})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path_match_regex")
	})

	t.Run("non-recursive ProtoCollection", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "proto-collection-non-recursive-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create a file in the root
		err = os.WriteFile(filepath.Join(tempDir, "root.proto"), []byte(`syntax = "proto3";`), 0o600)
		require.NoError(t, err)

		// Create a file in a subdirectory
		subDir := filepath.Join(tempDir, "subdir")
		err = os.Mkdir(subDir, 0o750)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(subDir, "sub.proto"), []byte(`syntax = "proto3";`), 0o600)
		require.NoError(t, err)

		protoCollection := configv1.ProtoCollection_builder{
			RootPath:       &tempDir,
			PathMatchRegex: proto.String(`\.proto$`),
			IsRecursive:    proto.Bool(false),
		}.Build()

		fds, err := ParseProtoFromDefs(ctx, nil, []*configv1.ProtoCollection{protoCollection})
		require.NoError(t, err)
		require.Len(t, fds.File, 1)
		assert.Equal(t, "root.proto", fds.File[0].GetName()) // Should only find the root proto
	})

	t.Run("error on ProtoFile with no content or path", func(t *testing.T) {
		protoDef := configv1.ProtoDefinition_builder{
			ProtoFile: configv1.ProtoFile_builder{
				FileName: proto.String("empty.proto"),
			}.Build(),
		}.Build()
		_, err := ParseProtoFromDefs(ctx, []*configv1.ProtoDefinition{protoDef}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has neither content nor a path")
	})
}

func TestParseProtoByReflection_Extended(t *testing.T) {
	// Setup a mock server
	server := &mockReflectionServer{streamReady: make(chan struct{})}
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	s := grpc.NewServer()
	reflectpb.RegisterServerReflectionServer(s, server)
	go func() { _ = s.Serve(lis) }()
	defer s.Stop()

	t.Run("successful reflection", func(t *testing.T) {
		go func() {
			<-server.streamReady
			// Simulate the server sending a response
			req, err := server.stream.Recv()
			require.NoError(t, err)
			assert.NotNil(t, req.GetListServices())

			err = server.stream.Send(&reflectpb.ServerReflectionResponse{
				MessageResponse: &reflectpb.ServerReflectionResponse_ListServicesResponse{
					ListServicesResponse: &reflectpb.ListServiceResponse{
						Service: []*reflectpb.ServiceResponse{{Name: "test.Service"}},
					},
				},
			})
			require.NoError(t, err)

			// Simulate the server sending a response for FileContainingSymbol
			req, err = server.stream.Recv()
			require.NoError(t, err)
			assert.NotNil(t, req.GetFileContainingSymbol())

			fdp := &descriptorpb.FileDescriptorProto{
				Name: proto.String("test.proto"),
			}
			fdpBytes, err := proto.Marshal(fdp)
			require.NoError(t, err)

			err = server.stream.Send(&reflectpb.ServerReflectionResponse{
				MessageResponse: &reflectpb.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectpb.FileDescriptorResponse{
						FileDescriptorProto: [][]byte{fdpBytes},
					},
				},
			})
			require.NoError(t, err)
		}()
		_, err := ParseProtoByReflection(context.Background(), lis.Addr().String())
		require.NoError(t, err)
	})
}

func TestExtractMcpDefinitions_Extended(t *testing.T) {
	t.Run("extracts prompts and complex tools", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:    proto.String("test.proto"),
					Package: proto.String("test"),
					MessageType: []*descriptorpb.DescriptorProto{
						{Name: proto.String("ToolRequest")},
						{Name: proto.String("ToolResponse")},
						{
							Name:    proto.String("ResourceMsg"),
							Options: &descriptorpb.MessageOptions{},
						},
					},
					Service: []*descriptorpb.ServiceDescriptorProto{
						{
							Name: proto.String("TestService"),
							Method: []*descriptorpb.MethodDescriptorProto{
								{
									Name:       proto.String("MyTool"),
									InputType:  proto.String(".test.ToolRequest"),
									OutputType: proto.String(".test.ToolResponse"),
									Options:    &descriptorpb.MethodOptions{},
								},
							},
						},
					},
				},
			},
		}

		// Add MCP options dynamically
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_ToolName, "MyAwesomeTool")
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_ToolDescription, "An awesome tool.")
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_McpToolReadonlyHint, true)
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_McpToolDestructiveHint, true)
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_McpToolIdempotentHint, true)
		proto.SetExtension(fds.File[0].Service[0].Method[0].Options, mcpopt.E_McpToolOpenworldHint, true)
		proto.SetExtension(fds.File[0].MessageType[2].Options, mcpopt.E_ResourceName, "MyResource")
		proto.SetExtension(fds.File[0].MessageType[2].Options, mcpopt.E_ResourceDescription, "A test resource.")

		parsedData, err := ExtractMcpDefinitions(fds)
		require.NoError(t, err)
		assert.NotNil(t, parsedData)

		// Check Tools
		require.Len(t, parsedData.Tools, 1)
		tool := parsedData.Tools[0]
		assert.Equal(t, "MyAwesomeTool", tool.Name)
		assert.Equal(t, "An awesome tool.", tool.Description)
		assert.True(t, tool.ReadOnlyHint)
		assert.True(t, tool.DestructiveHint)
		assert.True(t, tool.IdempotentHint)
		assert.True(t, tool.OpenWorldHint)

		// Check Resources
		require.Len(t, parsedData.Resources, 1)
		resource := parsedData.Resources[0]
		assert.Equal(t, "MyResource", resource.Name)
		assert.Equal(t, "A test resource.", resource.Description)
		assert.Equal(t, "test.ResourceMsg", resource.MessageType)
	})

	t.Run("method with no mcp annotations", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:    proto.String("test.proto"),
					Package: proto.String("test"),
					MessageType: []*descriptorpb.DescriptorProto{
						{Name: proto.String("ToolRequest")},
						{Name: proto.String("ToolResponse")},
					},
					Service: []*descriptorpb.ServiceDescriptorProto{
						{
							Name: proto.String("TestService"),
							Method: []*descriptorpb.MethodDescriptorProto{
								{
									Name:       proto.String("MyTool"),
									InputType:  proto.String(".test.ToolRequest"),
									OutputType: proto.String(".test.ToolResponse"),
								},
							},
						},
					},
				},
			},
		}

		parsedData, err := ExtractMcpDefinitions(fds)
		require.NoError(t, err)
		assert.NotNil(t, parsedData)
		require.Len(t, parsedData.Tools, 1)
		// Default tool name should be the method name
		assert.Equal(t, "MyTool", parsedData.Tools[0].Name)
		assert.Empty(t, parsedData.Tools[0].Description)
	})
}
