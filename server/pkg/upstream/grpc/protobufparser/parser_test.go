package protobufparser

import (
	"errors"
	"os"
	"testing"

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
}
