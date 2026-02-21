package filesystem

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockZeroSizeFs wraps an afero.Fs and mocks Stat() to return 0 size for files.
// This simulates special files like /proc/self/environ or /dev/zero which report 0 size
// but contain data.
type MockZeroSizeFs struct {
	afero.Fs
}

func (m *MockZeroSizeFs) Stat(name string) (os.FileInfo, error) {
	info, err := m.Fs.Stat(name)
	if err != nil {
		return nil, err
	}
	return &MockZeroSizeFileInfo{FileInfo: info}, nil
}

func (m *MockZeroSizeFs) Open(name string) (afero.File, error) {
	// We don't need to wrap the file because readFileTool uses io.ReadAll(reader)
	// and the reader comes from the file returned by Open.
	// However, readFileTool calls fs.Open(resolvedPath).
	// If we use MockZeroSizeFs as the fs passed to readFileTool, fs.Open calls m.Fs.Open.
	return m.Fs.Open(name)
}

type MockZeroSizeFileInfo struct {
	os.FileInfo
}

func (m *MockZeroSizeFileInfo) Size() int64 {
	return 0
}

func TestReadFileTool_DoSPrevention(t *testing.T) {
	// Create a base filesystem with a large file
	baseFs := afero.NewMemMapFs()
	largeFilePath := "/large_file.txt"

	// Create a file slightly larger than the limit (10MB)
	// Limit is 10 * 1024 * 1024 = 10485760 bytes
	limit := 10 * 1024 * 1024
	largeData := make([]byte, limit+100)
	// Fill with some data
	for i := range largeData {
		largeData[i] = 'A'
	}

	err := afero.WriteFile(baseFs, largeFilePath, largeData, 0644)
	require.NoError(t, err)

	// Wrap with MockZeroSizeFs
	mockFs := &MockZeroSizeFs{Fs: baseFs}

	// Create provider that uses this mock fs
	// We need a provider that returns our mock fs
	prov := &dosMockProvider{fs: mockFs}

	// Create tool definition
	toolDef := readFileTool(prov, mockFs)
	handler := toolDef.Handler

	// Execute tool
	args := map[string]interface{}{
		"path": largeFilePath,
	}

	result, err := handler(context.Background(), args)

	// Expect error due to size limit exceeded, despite Stat() reporting 0 size
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("file size exceeds limit of %d bytes", limit))
	assert.Nil(t, result)
}

// dosMockProvider implements provider.Provider interface
type dosMockProvider struct {
	fs afero.Fs
}

func (m *dosMockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *dosMockProvider) ResolvePath(path string) (string, error) {
	return path, nil
}

func (m *dosMockProvider) Close() error {
	return nil
}
