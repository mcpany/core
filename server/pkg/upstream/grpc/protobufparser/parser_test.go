package protobufparser

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcpopt "github.com/mcpany/core/proto/mcp_options/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const testProtoFilename = "test.proto"

// MockServerReflectionStream is a mock implementation of the ServerReflection_ServerReflectionInfoClient
type MockServerReflectionStream struct {
	mock.Mock
	grpc.ClientStream
}

func (m *MockServerReflectionStream) Send(req *reflectpb.ServerReflectionRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockServerReflectionStream) Recv() (*reflectpb.ServerReflectionResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reflectpb.ServerReflectionResponse), args.Error(1)
}

func (m *MockServerReflectionStream) CloseSend() error {
	args := m.Called()
	return args.Error(0)
}

func TestGetFileDescriptorByFilename(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockStream := new(MockServerReflectionStream)
		fdp := &descriptorpb.FileDescriptorProto{
			Name: proto.String(testProtoFilename),
		}
		fdpBytes, err := proto.Marshal(fdp)
		require.NoError(t, err)

		req := &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
				FileByFilename: testProtoFilename,
			},
		}
		resp := &reflectpb.ServerReflectionResponse{
			MessageResponse: &reflectpb.ServerReflectionResponse_FileDescriptorResponse{
				FileDescriptorResponse: &reflectpb.FileDescriptorResponse{
					FileDescriptorProto: [][]byte{fdpBytes},
				},
			},
		}

		mockStream.On("Send", req).Return(nil)
		mockStream.On("Recv").Return(resp, nil)

		resultFdp, err := getFileDescriptorByFilename(mockStream, testProtoFilename)
		require.NoError(t, err)
		assert.True(t, proto.Equal(fdp, resultFdp), "The returned FileDescriptorProto should match the expected one.")

		mockStream.AssertExpectations(t)
	})

	t.Run("send error", func(t *testing.T) {
		mockStream := new(MockServerReflectionStream)
		sendErr := errors.New("send error")

		req := &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
				FileByFilename: testProtoFilename,
			},
		}

		mockStream.On("Send", req).Return(sendErr)

		_, err := getFileDescriptorByFilename(mockStream, testProtoFilename)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), sendErr.Error())

		mockStream.AssertExpectations(t)
	})

	t.Run("receive error", func(t *testing.T) {
		mockStream := new(MockServerReflectionStream)
		recvErr := errors.New("receive error")

		req := &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
				FileByFilename: testProtoFilename,
			},
		}

		mockStream.On("Send", req).Return(nil)
		mockStream.On("Recv").Return(nil, recvErr)

		_, err := getFileDescriptorByFilename(mockStream, testProtoFilename)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), recvErr.Error())

		mockStream.AssertExpectations(t)
	})

	t.Run("empty response", func(t *testing.T) {
		mockStream := new(MockServerReflectionStream)

		req := &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
				FileByFilename: testProtoFilename,
			},
		}
		resp := &reflectpb.ServerReflectionResponse{
			MessageResponse: &reflectpb.ServerReflectionResponse_FileDescriptorResponse{
				FileDescriptorResponse: &reflectpb.FileDescriptorResponse{
					FileDescriptorProto: [][]byte{}, // Empty
				},
			},
		}

		mockStream.On("Send", req).Return(nil)
		mockStream.On("Recv").Return(resp, nil)

		_, err := getFileDescriptorByFilename(mockStream, testProtoFilename)
		assert.Error(t, err)

		mockStream.AssertExpectations(t)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		mockStream := new(MockServerReflectionStream)
		invalidBytes := []byte("invalid proto bytes")

		req := &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
				FileByFilename: testProtoFilename,
			},
		}
		resp := &reflectpb.ServerReflectionResponse{
			MessageResponse: &reflectpb.ServerReflectionResponse_FileDescriptorResponse{
				FileDescriptorResponse: &reflectpb.FileDescriptorResponse{
					FileDescriptorProto: [][]byte{invalidBytes},
				},
			},
		}

		mockStream.On("Send", req).Return(nil)
		mockStream.On("Recv").Return(resp, nil)

		_, err := getFileDescriptorByFilename(mockStream, testProtoFilename)
		assert.Error(t, err)

		mockStream.AssertExpectations(t)
	})
}

func loadTestFileDescriptorSet(t *testing.T) *descriptorpb.FileDescriptorSet {
	t.Helper()
	// This path is relative to the package directory where the test is run.
	b, err := os.ReadFile("../../../../../build/all.protoset")
	require.NoError(t, err, "Failed to read protoset file. Ensure 'make gen' has been run.")

	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(b, fds)
	require.NoError(t, err, "Failed to unmarshal protoset file")

	return fds
}

func TestExtractMcpDefinitions(t *testing.T) {
	fds := loadTestFileDescriptorSet(t)

	t.Run("successful extraction", func(t *testing.T) {
		parsedData, err := ExtractMcpDefinitions(fds)
		require.NoError(t, err)
		assert.NotNil(t, parsedData)

		// Basic checks
		assert.NotEmpty(t, parsedData.Tools)

		// Find a specific tool to inspect
		var getWeatherTool *McpTool
		for i, tool := range parsedData.Tools {
			if tool.Name == "GetWeather" {
				getWeatherTool = &parsedData.Tools[i]
				break
			}
		}

		require.NotNil(t, getWeatherTool, "Tool 'GetWeather' should be found")
		assert.Equal(t, "", getWeatherTool.Description)
		assert.Equal(t, "WeatherService", getWeatherTool.ServiceName)
		assert.Equal(t, "GetWeather", getWeatherTool.MethodName)
		assert.Equal(t, "/examples.weather.v1.WeatherService/GetWeather", getWeatherTool.FullMethodName)
		assert.Equal(t, "examples.weather.v1.GetWeatherRequest", getWeatherTool.RequestType)
		assert.Equal(t, "examples.weather.v1.GetWeatherResponse", getWeatherTool.ResponseType)
		assert.False(t, getWeatherTool.IdempotentHint)
		assert.False(t, getWeatherTool.DestructiveHint)

		// Check request fields
		require.Len(t, getWeatherTool.RequestFields, 1)
		assert.Equal(t, "location", getWeatherTool.RequestFields[0].Name)
		assert.Equal(t, "", getWeatherTool.RequestFields[0].Description)
		assert.Equal(t, "string", getWeatherTool.RequestFields[0].Type)
		assert.False(t, getWeatherTool.RequestFields[0].IsRepeated)
	})

	t.Run("nil fds", func(t *testing.T) {
		_, err := ExtractMcpDefinitions(nil)
		assert.Error(t, err)
	})

	t.Run("corrupted fds", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:       proto.String("invalid.proto"),
					Package:    proto.String("invalid"),
					Dependency: []string{"a", "b"},
				},
				{
					Name: proto.String("a"),
				},
			},
		}
		_, err := ExtractMcpDefinitions(fds)
		assert.Error(t, err, "Should fail with corrupted/incomplete FileDescriptorSet")
	})
}

func TestMcpField_Getters(t *testing.T) {
	field := McpField{
		Name:        "test_name",
		Description: "test_description",
		Type:        "test_type",
		IsRepeated:  true,
	}

	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "test_name", field.GetName())
	})

	t.Run("GetDescription", func(t *testing.T) {
		assert.Equal(t, "test_description", field.GetDescription())
	})

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, "test_type", field.GetType())
	})

	t.Run("GetIsRepeated", func(t *testing.T) {
		assert.True(t, field.GetIsRepeated())
	})
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

	t.Run("recursive ProtoCollection", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "proto-collection-recursive-*")
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
			IsRecursive:    proto.Bool(true),
		}.Build()

		fds, err := ParseProtoFromDefs(ctx, nil, []*configv1.ProtoCollection{protoCollection})
		require.NoError(t, err)
		require.Len(t, fds.File, 2)
		// Order is not guaranteed, so check map or existence
		names := make(map[string]bool)
		for _, f := range fds.File {
			names[f.GetName()] = true
		}
		assert.True(t, names["root.proto"])
		assert.True(t, names["subdir/sub.proto"])
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
	setupServer := func(t *testing.T) (*mockReflectionServer, *grpc.Server, net.Listener) {
		server := &mockReflectionServer{streamReady: make(chan struct{})}
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		s := grpc.NewServer()
		reflectpb.RegisterServerReflectionServer(s, server)
		go func() { _ = s.Serve(lis) }()
		return server, s, lis
	}

	t.Run("successful reflection", func(t *testing.T) {
		server, s, lis := setupServer(t)
		defer s.Stop()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		wg.Wait()
	})

	t.Run("ListServices failure", func(t *testing.T) {
		server, s, lis := setupServer(t)
		defer s.Stop()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-server.streamReady
			// Simulate the server sending an error response or closing early
			server.stream.SendMsg(&reflectpb.ServerReflectionResponse{})
			// Sending an empty response (which is invalid for ListServices expectations if we check type)
			// But MockServerReflectionStream is just a wrapper around the real stream, so we can't easily inject error unless we mock the client.
			// However, here we are using a real gRPC client against a mock server.
			// So we can make the server return an error or invalid response.

			err := server.stream.Send(&reflectpb.ServerReflectionResponse{
				// Missing MessageResponse
			})
			require.NoError(t, err)
		}()
		_, err := ParseProtoByReflection(context.Background(), lis.Addr().String())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid response type")
		wg.Wait()
	})

	t.Run("FileContainingSymbol failure", func(t *testing.T) {
		server, s, lis := setupServer(t)
		defer s.Stop()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-server.streamReady
			// 1. ListServices success
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

			// 2. FileContainingSymbol failure
			req, err = server.stream.Recv()
			require.NoError(t, err)
			assert.NotNil(t, req.GetFileContainingSymbol())

			// Send error response (empty FileDescriptorProto list)
			err = server.stream.Send(&reflectpb.ServerReflectionResponse{
				MessageResponse: &reflectpb.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectpb.FileDescriptorResponse{
						FileDescriptorProto: [][]byte{}, // Empty
					},
				},
			})
			require.NoError(t, err)
		}()
		_, err := ParseProtoByReflection(context.Background(), lis.Addr().String())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or empty response")
		wg.Wait()
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
