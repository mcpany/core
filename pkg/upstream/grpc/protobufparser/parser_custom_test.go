/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package protobufparser

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestParseProtoFromDefs(t *testing.T) {
	t.Run("successful parsing with ProtoCollection", func(t *testing.T) {
		// Create a temporary directory and a sample proto file
		tempDir, err := os.MkdirTemp("", "test-proto-collection")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

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
		err = os.WriteFile(protoFilePath, []byte(protoContent), 0644)
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
