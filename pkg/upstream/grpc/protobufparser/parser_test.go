
package protobufparser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func loadTestFileDescriptorSet(t *testing.T) *descriptorpb.FileDescriptorSet {
	t.Helper()
	// This path is relative to the package directory where the test is run.
	b, err := os.ReadFile("../../../../build/all.protoset")
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
