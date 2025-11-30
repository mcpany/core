package filesystem

import (
	"context"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestFileSystemService(t *testing.T) {
	testDir, err := os.MkdirTemp("", "filesystem-test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	configFile := fmt.Sprintf(`
upstreamServices:
  - name: "my-filesystem-service"
    filesystemService:
      rootDirectory: "%s"
`, testDir)

	testFilePath := "test.txt"
	testFileContent := "Hello, world!"

	s := integration.StartMCPANYServerWithConfig(t, "filesystem-test", configFile)
	defer s.CleanupFunc()

	// Write a file
	writeFileArgs, err := json.Marshal(map[string]interface{}{
		"path":    testFilePath,
		"content": testFileContent,
	})
	require.NoError(t, err)
	_, err = s.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "my-filesystem-service/writeFile",
		Arguments: writeFileArgs,
	})
	require.NoError(t, err)

	// Read the file
	readFileArgs, err := json.Marshal(map[string]interface{}{
		"path": testFilePath,
	})
	require.NoError(t, err)
	resp, err := s.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "my-filesystem-service/readFile",
		Arguments: readFileArgs,
	})
	require.NoError(t, err)
	require.Equal(t, testFileContent, resp.Content[0].(*mcp.TextContent).Text)

	// List the files
	listFileArgs, err := json.Marshal(map[string]interface{}{
		"path": ".",
	})
	require.NoError(t, err)
	resp, err = s.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "my-filesystem-service/listFiles",
		Arguments: listFileArgs,
	})
	require.NoError(t, err)
	var list []interface{}
	err = json.Unmarshal([]byte(resp.Content[0].(*mcp.TextContent).Text), &list)
	require.NoError(t, err)
	require.Contains(t, list, testFilePath)
}
