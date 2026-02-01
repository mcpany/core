// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSearchFiles_Comprehensive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "search_comprehensive")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_comp"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/data": tempDir,
			},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	findTool := func(name string) tool.Tool {
		tTool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tTool
		}
		return nil
	}
	search := findTool("search_files")
	require.NotNil(t, search)

	t.Run("ExcludePatterns", func(t *testing.T) {
		// Create a secret file
		secretPath := filepath.Join(tempDir, "secret.key")
		err := os.WriteFile(secretPath, []byte("my-secret-key"), 0644)
		require.NoError(t, err)

		// Create a normal file
		normalPath := filepath.Join(tempDir, "public.txt")
		err = os.WriteFile(normalPath, []byte("public-info"), 0644)
		require.NoError(t, err)

		// Search matching both, but excluding .key
		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path": "/data",
				"pattern": ".*",
				"exclude_patterns": []interface{}{"*.key"},
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should find public.txt but NOT secret.key
		foundPublic := false
		for _, m := range matches {
			f := filepath.Base(m["file"].(string))
			if f == "secret.key" {
				t.Error("Should have excluded secret.key")
			}
			if f == "public.txt" {
				foundPublic = true
			}
		}
		assert.True(t, foundPublic, "Should have found public.txt")
	})

	t.Run("HiddenDirectories", func(t *testing.T) {
		// Create hidden directory
		hiddenDir := filepath.Join(tempDir, ".hidden")
		err := os.Mkdir(hiddenDir, 0755)
		require.NoError(t, err)

		// Create file inside hidden directory
		hiddenFile := filepath.Join(hiddenDir, "config.json")
		err = os.WriteFile(hiddenFile, []byte("hidden-content"), 0644)
		require.NoError(t, err)

		// Create file in root
		rootFile := filepath.Join(tempDir, "visible.txt")
		err = os.WriteFile(rootFile, []byte("visible-content"), 0644)
		require.NoError(t, err)

		// Search everything
		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path": "/data",
				"pattern": ".*",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should find visible.txt but NOT config.json
		foundVisible := false
		for _, m := range matches {
			f := m["file"].(string)
			if filepath.Base(f) == "config.json" {
				t.Error("Should have skipped file in hidden directory")
			}
			if filepath.Base(f) == "visible.txt" {
				foundVisible = true
			}
		}
		assert.True(t, foundVisible, "Should have found visible.txt")
	})

	t.Run("LargeFiles", func(t *testing.T) {
		// Create a large file > 10MB
		largePath := filepath.Join(tempDir, "large.dat")
		f, err := os.Create(largePath)
		require.NoError(t, err)

		// Write 10MB + 1 byte
		// Writing in chunks to avoid memory issues in test
		chunk := make([]byte, 1024*1024) // 1MB
		for i := 0; i < 10; i++ {
			f.Write(chunk)
		}
		f.Write([]byte{0})
		f.Close()

		// Create a small file
		smallPath := filepath.Join(tempDir, "small.dat")
		err = os.WriteFile(smallPath, []byte("small-content"), 0644)
		require.NoError(t, err)

		// Search
		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path": "/data",
				"pattern": ".*",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should find small.dat but NOT large.dat
		foundSmall := false
		for _, m := range matches {
			f := filepath.Base(m["file"].(string))
			if f == "large.dat" {
				t.Error("Should have skipped large file")
			}
			if f == "small.dat" {
				foundSmall = true
			}
		}
		assert.True(t, foundSmall, "Should have found small.dat")
	})
}
