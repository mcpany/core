
package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestClient_GetTools(t *testing.T) {
	client := NewClient("/tmp")
	tools, err := client.GetTools(context.Background())
	assert.NoError(t, err)
	assert.Len(t, tools, 4)
}

func TestClient_CallTool(t *testing.T) {
	basePath := t.TempDir()
	client := NewClient(basePath)

	// Test WriteFile
	_, err := client.CallTool(context.Background(), WriteFileToolName, &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"path":    {Kind: &structpb.Value_StringValue{StringValue: "test.txt"}},
			"content": {Kind: &structpb.Value_StringValue{StringValue: "Hello, world!"}},
		},
	})
	assert.NoError(t, err)

	// Test ReadFile
	result, err := client.CallTool(context.Background(), ReadFileToolName, &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"path": {Kind: &structpb.Value_StringValue{StringValue: "test.txt"}},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", result.Fields["content"].GetStringValue())

	// Test ListFiles
	result, err = client.CallTool(context.Background(), ListFilesToolName, &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"path": {Kind: &structpb.Value_StringValue{StringValue: "."}},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, result.Fields["files"].GetListValue().Values, 1)
	assert.Equal(t, "test.txt", result.Fields["files"].GetListValue().Values[0].GetStringValue())

	// Test DeleteFile
	_, err = client.CallTool(context.Background(), DeleteFileToolName, &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"path": {Kind: &structpb.Value_StringValue{StringValue: "test.txt"}},
		},
	})
	assert.NoError(t, err)

	// Verify that the file was deleted
	_, err = os.Stat(filepath.Join(basePath, "test.txt"))
	assert.True(t, os.IsNotExist(err))
}
