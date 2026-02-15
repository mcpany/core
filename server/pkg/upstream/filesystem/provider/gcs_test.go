package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestNewGcsProvider(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		_, err := NewGcsProvider(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gcs config is nil")
	})

	t.Run("Valid Config", func(t *testing.T) {
		// This test relies on default client behavior, which might fail without creds.
		// That's fine, we preserve existing behavior.
		config := configv1.GcsFs_builder{
			Bucket: proto.String("my-bucket"),
		}.Build()
		p, err := NewGcsProvider(context.Background(), config)
		if err == nil {
			defer p.Close()
			assert.NotNil(t, p)
			assert.NotNil(t, p.GetFs())
		} else {
			// It might fail if no creds are found, which is acceptable in this env.
			t.Logf("NewGcsProvider failed as expected (no creds): %v", err)
		}
	})
}

func TestGcsProvider_WithMockClient(t *testing.T) {
	// Backup and restore newStorageClient
	oldNewStorageClient := newStorageClient
	defer func() { newStorageClient = oldNewStorageClient }()

	newStorageClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		// Create a client with a mock HTTP client
		httpClient := &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Dummy response
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("content")),
						Header:     make(http.Header),
					}, nil
				},
			},
		}
		return storage.NewClient(ctx, option.WithHTTPClient(httpClient))
	}

	config := configv1.GcsFs_builder{
		Bucket: proto.String("my-bucket"),
	}.Build()

	p, err := NewGcsProvider(context.Background(), config)
	require.NoError(t, err)
	defer p.Close()

	// Test GetFs
	fs := p.GetFs()
	assert.NotNil(t, fs)
	assert.Equal(t, "gcs", fs.Name())

	// Test unsupported methods (covered by unit test, but good to verify on instance)
	err = fs.Mkdir("dir", 0755)
	assert.NoError(t, err)

	err = fs.MkdirAll("dir/subdir", 0755)
	assert.NoError(t, err)

	err = fs.Chmod("file", 0644)
	assert.NoError(t, err)
}

func TestNewGcsProvider_ClientError(t *testing.T) {
	// Backup and restore newStorageClient
	oldNewStorageClient := newStorageClient
	defer func() { newStorageClient = oldNewStorageClient }()

	newStorageClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		return nil, fmt.Errorf("mock error")
	}

	config := configv1.GcsFs_builder{
		Bucket: proto.String("my-bucket"),
	}.Build()

	_, err := NewGcsProvider(context.Background(), config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gcs client")
	assert.Contains(t, err.Error(), "mock error")
}

func TestGcsProvider_ResolvePath(t *testing.T) {
	// We can manually create a GcsProvider struct to test ResolvePath without needing valid creds
	p := &GcsProvider{}

	tests := []struct {
		name        string
		virtualPath string
		want        string
		wantErr     bool
	}{
		{
			name:        "Root",
			virtualPath: "/",
			wantErr:     true,
		},
		{
			name:        "Empty",
			virtualPath: "",
			wantErr:     true,
		},
		{
			name:        "Normal file",
			virtualPath: "/path/to/file.txt",
			want:        "path/to/file.txt",
			wantErr:     false,
		},
		{
			name:        "Traversal",
			virtualPath: "../file.txt",
			want:        "file.txt",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ResolvePath(tt.virtualPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGcsFileInfo(t *testing.T) {
	fi := &gcsFileInfo{
		name:    "test.txt",
		size:    123,
		modTime: time.Now(),
		isDir:   false,
	}

	assert.Equal(t, "test.txt", fi.Name())
	assert.Equal(t, int64(123), fi.Size())
	assert.Equal(t, os.FileMode(0644), fi.Mode())
	assert.False(t, fi.IsDir())
	assert.Nil(t, fi.Sys())
	assert.NotZero(t, fi.ModTime())

	dirFi := &gcsFileInfo{
		name:  "dir",
		isDir: true,
	}
	assert.Equal(t, os.ModeDir|0755, dirFi.Mode())
	assert.True(t, dirFi.IsDir())
}

func TestGcsFs_Methods(t *testing.T) {
	var fs afero.Fs = &gcsFs{}
	// Just verify it compiles and implements the interface
	assert.NotNil(t, fs)
	assert.Equal(t, "gcs", fs.Name())

	// These return nil/error without client, so we can test the "Not Supported" ones
	assert.NoError(t, fs.Chmod("foo", 0))
	assert.NoError(t, fs.Chown("foo", 0, 0))
	assert.NoError(t, fs.Chtimes("foo", time.Now(), time.Now()))
}

func TestGcsFile_Methods_Errors(t *testing.T) {
	f := &gcsFile{}

	_, err := f.Read([]byte{})
	assert.Error(t, err)
	assert.Equal(t, "file not opened for reading", err.Error())

	_, err = f.Write([]byte{})
	assert.Error(t, err)
	assert.Equal(t, "file not opened for writing", err.Error())

	_, err = f.Seek(0, 0)
	assert.Error(t, err)
	assert.Equal(t, "seek not supported", err.Error())

	_, err = f.WriteAt([]byte{}, 0)
	assert.Error(t, err)
	assert.Equal(t, "writeat not supported", err.Error())

	err = f.Truncate(0)
	assert.Error(t, err)
	assert.Equal(t, "truncate not supported", err.Error())

	// Sync does nothing
	assert.NoError(t, f.Sync())

	// Close on empty file
	assert.NoError(t, f.Close())
}
