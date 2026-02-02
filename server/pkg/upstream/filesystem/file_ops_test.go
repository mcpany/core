// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProvider is a mock implementation of provider.Provider
type MockProvider struct {
	mock.Mock
	fs afero.Fs
}

func (m *MockProvider) GetFs() afero.Fs {
	args := m.Called()
	if fs, ok := args.Get(0).(afero.Fs); ok {
		return fs
	}
	return m.fs
}

func (m *MockProvider) ResolvePath(virtualPath string) (string, error) {
	args := m.Called(virtualPath)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestReadFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	mockProv := new(MockProvider)
	mockProv.fs = fs

	// Setup initial state
	require.NoError(t, afero.WriteFile(fs, "/safe/hello.txt", []byte("hello world"), 0644))
	require.NoError(t, fs.MkdirAll("/safe/dir", 0755))

	tool := readFileTool(mockProv, fs)

	tests := []struct {
		name      string
		args      map[string]interface{}
		setupMock func()
		wantErr   bool
		wantCheck func(t *testing.T, res map[string]interface{})
	}{
		{
			name: "Happy Path",
			args: map[string]interface{}{"path": "hello.txt"},
			setupMock: func() {
				mockProv.On("ResolvePath", "hello.txt").Return("/safe/hello.txt", nil).Once()
			},
			wantErr: false,
			wantCheck: func(t *testing.T, res map[string]interface{}) {
				assert.Equal(t, "hello world", res["content"])
			},
		},
		{
			name: "File Not Found",
			args: map[string]interface{}{"path": "missing.txt"},
			setupMock: func() {
				mockProv.On("ResolvePath", "missing.txt").Return("/safe/missing.txt", nil).Once()
			},
			wantErr: true,
		},
		{
			name: "Path is Directory",
			args: map[string]interface{}{"path": "dir"},
			setupMock: func() {
				mockProv.On("ResolvePath", "dir").Return("/safe/dir", nil).Once()
			},
			wantErr: true,
		},
		{
			name: "Resolve Error",
			args: map[string]interface{}{"path": "../unsafe"},
			setupMock: func() {
				mockProv.On("ResolvePath", "../unsafe").Return("", fmt.Errorf("path traversal detected")).Once()
			},
			wantErr: true,
		},
		{
			name:    "Missing Path Arg",
			args:    map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}
			res, err := tool.Handler(context.Background(), tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantCheck != nil {
					tt.wantCheck(t, res)
				}
			}
			mockProv.AssertExpectations(t)
		})
	}
}

func TestWriteFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	mockProv := new(MockProvider)
	mockProv.fs = fs

	tests := []struct {
		name      string
		readOnly  bool
		args      map[string]interface{}
		setupMock func()
		wantErr   bool
		verify    func(t *testing.T)
	}{
		{
			name:     "Happy Path",
			readOnly: false,
			args: map[string]interface{}{
				"path":    "new.txt",
				"content": "new content",
			},
			setupMock: func() {
				mockProv.On("ResolvePath", "new.txt").Return("/safe/new.txt", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				content, err := afero.ReadFile(fs, "/safe/new.txt")
				assert.NoError(t, err)
				assert.Equal(t, "new content", string(content))
			},
		},
		{
			name:     "Read Only",
			readOnly: true,
			args: map[string]interface{}{
				"path":    "new.txt",
				"content": "content",
			},
			wantErr: true,
		},
		{
			name:     "Create Directory",
			readOnly: false,
			args: map[string]interface{}{
				"path":    "nested/dir/file.txt",
				"content": "nested",
			},
			setupMock: func() {
				mockProv.On("ResolvePath", "nested/dir/file.txt").Return("/safe/nested/dir/file.txt", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				content, err := afero.ReadFile(fs, "/safe/nested/dir/file.txt")
				assert.NoError(t, err)
				assert.Equal(t, "nested", string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := writeFileTool(mockProv, fs, tt.readOnly)
			if tt.setupMock != nil {
				tt.setupMock()
			}
			_, err := tool.Handler(context.Background(), tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t)
				}
			}
			mockProv.AssertExpectations(t)
		})
	}
}

func TestDeleteFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	mockProv := new(MockProvider)
	mockProv.fs = fs

	tests := []struct {
		name      string
		readOnly  bool
		setupFs   func()
		args      map[string]interface{}
		setupMock func()
		wantErr   bool
		verify    func(t *testing.T)
	}{
		{
			name:     "Delete File",
			readOnly: false,
			setupFs: func() {
				_ = afero.WriteFile(fs, "/safe/del.txt", []byte("bye"), 0644)
			},
			args: map[string]interface{}{"path": "del.txt"},
			setupMock: func() {
				mockProv.On("ResolvePath", "del.txt").Return("/safe/del.txt", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				exists, _ := afero.Exists(fs, "/safe/del.txt")
				assert.False(t, exists)
			},
		},
		{
			name:     "Recursive Delete",
			readOnly: false,
			setupFs: func() {
				_ = fs.MkdirAll("/safe/dir/subdir", 0755)
				_ = afero.WriteFile(fs, "/safe/dir/file.txt", []byte("content"), 0644)
			},
			args: map[string]interface{}{
				"path":      "dir",
				"recursive": true,
			},
			setupMock: func() {
				mockProv.On("ResolvePath", "dir").Return("/safe/dir", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				exists, _ := afero.Exists(fs, "/safe/dir")
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFs != nil {
				tt.setupFs()
			}
			tool := deleteFileTool(mockProv, fs, tt.readOnly)
			if tt.setupMock != nil {
				tt.setupMock()
			}
			_, err := tool.Handler(context.Background(), tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t)
				}
			}
			mockProv.AssertExpectations(t)
		})
	}
}

func TestMoveFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	mockProv := new(MockProvider)
	mockProv.fs = fs

	tests := []struct {
		name      string
		setupFs   func()
		args      map[string]interface{}
		setupMock func()
		wantErr   bool
		verify    func(t *testing.T)
	}{
		{
			name: "Move File",
			setupFs: func() {
				_ = afero.WriteFile(fs, "/safe/src.txt", []byte("content"), 0644)
			},
			args: map[string]interface{}{
				"source":      "src.txt",
				"destination": "dest.txt",
			},
			setupMock: func() {
				mockProv.On("ResolvePath", "src.txt").Return("/safe/src.txt", nil).Once()
				mockProv.On("ResolvePath", "dest.txt").Return("/safe/dest.txt", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				existsSrc, _ := afero.Exists(fs, "/safe/src.txt")
				assert.False(t, existsSrc)
				content, err := afero.ReadFile(fs, "/safe/dest.txt")
				assert.NoError(t, err)
				assert.Equal(t, "content", string(content))
			},
		},
		{
			name: "Move to New Directory",
			setupFs: func() {
				_ = afero.WriteFile(fs, "/safe/src2.txt", []byte("content"), 0644)
			},
			args: map[string]interface{}{
				"source":      "src2.txt",
				"destination": "newdir/dest.txt",
			},
			setupMock: func() {
				mockProv.On("ResolvePath", "src2.txt").Return("/safe/src2.txt", nil).Once()
				mockProv.On("ResolvePath", "newdir/dest.txt").Return("/safe/newdir/dest.txt", nil).Once()
			},
			wantErr: false,
			verify: func(t *testing.T) {
				content, err := afero.ReadFile(fs, "/safe/newdir/dest.txt")
				assert.NoError(t, err)
				assert.Equal(t, "content", string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFs != nil {
				tt.setupFs()
			}
			tool := moveFileTool(mockProv, fs, false)
			if tt.setupMock != nil {
				tt.setupMock()
			}
			_, err := tool.Handler(context.Background(), tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t)
				}
			}
			mockProv.AssertExpectations(t)
		})
	}
}

// Ensure the size limit test for readFileTool
func TestReadFileTool_SizeLimit(t *testing.T) {
	fs := afero.NewMemMapFs()
	mockProv := new(MockProvider)
	mockProv.fs = fs

	// Create large file (11MB)
	largeContent := make([]byte, 11*1024*1024)
	_ = afero.WriteFile(fs, "/safe/large.bin", largeContent, 0644)

	tool := readFileTool(mockProv, fs)
	args := map[string]interface{}{"path": "large.bin"}

	mockProv.On("ResolvePath", "large.bin").Return("/safe/large.bin", nil).Once()

	_, err := tool.Handler(context.Background(), args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size exceeds limit")
	mockProv.AssertExpectations(t)
}
