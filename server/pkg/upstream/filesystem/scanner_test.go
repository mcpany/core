package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream_SearchFiles_ScannerError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_repro_scanner")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file with a very long line (> 64KB)
	// bufio.Scanner default max token size is 64 * 1024
	longLine := strings.Repeat("a", 64*1024+10) + "findme"
	testFile := filepath.Join(tempDir, "longline.txt")
	err = os.WriteFile(testFile, []byte(longLine), 0644)
	require.NoError(t, err)

	// Configure the upstream
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_scanner"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(false),
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// Helper to find a tool by name
	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	t.Run("search_files_long_line_missing_check", func(t *testing.T) {
		searchTool := findTool("search_files")
		require.NotNil(t, searchTool)

		// Search for "findme" which is at the end of the long line
		// The scanner should fail with ErrTooLong and stop scanning.
		// The current implementation ignores scanner.Err(), so it returns no matches (or partial matches if any were found before)
		// and NO error.
		// The desired behavior is either to return an error or handle long lines.
		// For this bug report, I am demonstrating that it fails silently.

		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/data",
				"pattern": "findme",
			},
		})

		// Currently, it returns no error and empty matches because scanner stops silently.
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// If the bug exists, matches will be empty because "findme" is after the break.
		// If fixed (e.g., by increasing buffer or returning error), this assertion will change.
		// Wait, if I want to PROVE the bug, I should assert that it FAILS to find it,
		// but ideally it SHOULD return an error saying "file too complex" or similar.

		// So the test essentially passes if the bug is present (it finds nothing).
		// But a "reproduction test case" should fail when the bug is present if I assert the CORRECT behavior.
		// The correct behavior is that it should return an error.

		// So I will assert that we get an error.
		// This assertion will FAIL on the current codebase.
		// But wait, the task says "Write a new test case that specifically fails before your fix and passes after it".
		// So I should assert that err is NOT nil (we expect an error for line too long if we handle it properly,
		// OR we expect the match to be found if we fix the scanner buffer).

		// If I decide to fix it by checking error, then `err` should be non-nil.
		// If I decide to fix it by increasing buffer, then `matches` should contain the item.

		// Let's assume the fix is to Report the error.
		// So checking `scanner.Err()` and returning it.
		// In that case, `searchTool.Execute` should return error.

		// However, `search_files` might just skip the file?
		// `Walk` function returns error. If it returns error, `afero.Walk` stops?
		// "If the function returns an error, Walk returns it; the only exception is SkipDir".
		// So yes, it should bubble up.

		// But wait, if I fix it to just check error, the test will pass (err != nil).
		// But right now, err == nil.
		// So Assert(err != nil) or Assert(len(matches) == 1).

		// Let's try to Assert(len(matches) == 1) implies I want to support long lines.
		// Let's try to Assert(err != nil) implies I want to fail on long lines.

		// Failing on long lines is safer and easier. Supporting arbitrary long lines requires `bufio.Reader` or custom buffer.
		// So I will fix it by reporting the error.

		// Therefore, the test expects an error.
		if len(matches) == 0 {
             // If we didn't find it, we MUST have an error explaining why.
             assert.Error(t, err, "Expected error due to long line, but got success with no matches")
        }
	})
}
