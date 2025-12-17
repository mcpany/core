// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package protobufparser

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestParseProtoFromDefs(t *testing.T) {
	t.Run("successful parsing with ProtoCollection", func(t *testing.T) {
		// Create a temporary directory and a sample proto file
		tempDir, err := os.MkdirTemp("", "test-proto-collection")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		protoContent := `
syntax = "proto3";
package test;

service TestService {
  rpc TestMethod(TestRequest) returns (TestResponse);
}

message TestRequest {
  string name = 1;
}

message TestResponse {
  string message = 1;
}
`
		protoFilePath := filepath.Join(tempDir, "test.proto")
		err = os.WriteFile(protoFilePath, []byte(protoContent), 0o600)
		require.NoError(t, err)

		// Create a ProtoCollection
		protoCollection := configv1.ProtoCollection_builder{
			RootPath:       &tempDir,
			PathMatchRegex: proto.String(".*\\.proto"),
			IsRecursive:    proto.Bool(true),
		}.Build()

		// Call ParseProtoFromDefs
		fds, err := ParseProtoFromDefs(context.Background(), nil, []*configv1.ProtoCollection{protoCollection})
		require.NoError(t, err)
		assert.NotNil(t, fds)

		// Check the parsed data
		require.Len(t, fds.File, 1)
		assert.Equal(t, "test.proto", fds.File[0].GetName())
		require.Len(t, fds.File[0].Service, 1)
		assert.Equal(t, "TestService", fds.File[0].Service[0].GetName())
	})

	t.Run("successful parsing with ProtoDefinition file_content", func(t *testing.T) {
		protoContent := `
syntax = "proto3";
package test2;

service TestService2 {
  rpc TestMethod2(TestRequest2) returns (TestResponse2);
}

message TestRequest2 {
  string name = 1;
}

message TestResponse2 {
  string message = 1;
}
`
		protoDefinitions := []*configv1.ProtoDefinition{
			configv1.ProtoDefinition_builder{
				ProtoFile: configv1.ProtoFile_builder{
					FileName:    proto.String("test2.proto"),
					FileContent: proto.String(protoContent),
				}.Build(),
			}.Build(),
		}

		// Call ParseProtoFromDefs
		fds, err := ParseProtoFromDefs(context.Background(), protoDefinitions, nil)
		require.NoError(t, err)
		assert.NotNil(t, fds)

		// Check the parsed data
		require.Len(t, fds.File, 1)
		assert.Equal(t, "test2.proto", fds.File[0].GetName())
		require.Len(t, fds.File[0].Service, 1)
		assert.Equal(t, "TestService2", fds.File[0].Service[0].GetName())
	})

	t.Run("no proto files", func(t *testing.T) {
		_, err := ParseProtoFromDefs(context.Background(), nil, nil)
		assert.Error(t, err)
	})
}

func TestProcessProtoCollection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "proto-collection-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	protoContent := `syntax = "proto3"; package test; message Test {}`
	protoFilePath := filepath.Join(tempDir, "test.proto")
	err = os.WriteFile(protoFilePath, []byte(protoContent), 0o600)
	require.NoError(t, err)

	collection := configv1.ProtoCollection_builder{
		RootPath:       &tempDir,
		PathMatchRegex: proto.String(".*\\.proto"),
		IsRecursive:    proto.Bool(true),
	}.Build()

	files, err := processProtoCollection(collection, tempDir)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, filepath.Join(tempDir, "test.proto"), files[0])
}

func TestWriteProtoFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "proto-file-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("from_content", func(t *testing.T) {
		protoFile := configv1.ProtoFile_builder{
			FileName:    proto.String("test.proto"),
			FileContent: proto.String(`syntax = "proto3"; package test; message Test {}`),
		}.Build()
		filePath, err := writeProtoFile(protoFile, tempDir)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, "test.proto"), filePath)

		content, err := os.ReadFile(filePath) //nolint:gosec // Test file
		require.NoError(t, err)
		assert.Equal(t, protoFile.GetFileContent(), string(content))
	})

	t.Run("from_path", func(t *testing.T) {
		protoContent := `syntax = "proto3"; package test; message Test {}`
		protoFilePath := filepath.Join(tempDir, "source.proto")
		err = os.WriteFile(protoFilePath, []byte(protoContent), 0o600)
		require.NoError(t, err)

		protoFile := configv1.ProtoFile_builder{
			FileName: proto.String("test.proto"),
			FilePath: proto.String(protoFilePath),
		}.Build()
		filePath, err := writeProtoFile(protoFile, tempDir)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, "test.proto"), filePath)

		content, err := os.ReadFile(filePath) //nolint:gosec // Test file
		require.NoError(t, err)
		assert.Equal(t, protoContent, string(content))
	})
}
