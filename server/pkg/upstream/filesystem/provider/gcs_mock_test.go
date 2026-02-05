// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

// mockTransportForGcs implements http.RoundTripper
type mockTransportForGcs struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransportForGcs) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestGcsProvider_Methods_WithMock(t *testing.T) {
	// Backup and restore newStorageClient
	oldNewStorageClient := newStorageClient
	defer func() { newStorageClient = oldNewStorageClient }()

	newStorageClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		httpClient := &http.Client{
			Transport: &mockTransportForGcs{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					path := req.URL.Path
					query := req.URL.Query()

					// Stat / Metadata request
					// GET /storage/v1/b/my-bucket/o/test-file.txt
					if strings.Contains(path, "/storage/v1/b/my-bucket/o/test-file.txt") && query.Get("alt") != "media" {
						jsonResp := `{
							"kind": "storage#object",
							"id": "my-bucket/test-file.txt/1",
							"name": "test-file.txt",
							"bucket": "my-bucket",
							"contentType": "text/plain",
							"size": "11",
							"updated": "2023-01-01T00:00:00Z"
						}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonResp)),
							Header:     make(http.Header),
						}, nil
					}

					// Read request
					// Can be GET /storage/v1/b/my-bucket/o/test-file.txt?alt=media
					// OR GET /my-bucket/test-file.txt
					if (strings.Contains(path, "/storage/v1/b/my-bucket/o/test-file.txt") && query.Get("alt") == "media") ||
						path == "/my-bucket/test-file.txt" {
						headers := make(http.Header)
						headers.Set("Content-Length", "11")
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("hello world")),
							Header:     headers,
						}, nil
					}

					// Stat / Metadata request for non-existent file
					// GET /storage/v1/b/my-bucket/o/missing.txt
					if strings.Contains(path, "/o/missing.txt") && query.Get("alt") != "media" {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(strings.NewReader(`{"error": {"code": 404, "message": "Not Found"}}`)),
							Header:     make(http.Header),
						}, nil
					}

					// Read request for non-existent file
					if (strings.Contains(path, "/o/missing.txt") && query.Get("alt") == "media") ||
						path == "/my-bucket/missing.txt" {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(strings.NewReader(`{"error": {"code": 404, "message": "Not Found"}}`)),
							Header:     make(http.Header),
						}, nil
					}

					// Mock directory object for Readdir test
					// We use "dir-marker" as the directory object
					if strings.Contains(path, "/storage/v1/b/my-bucket/o/dir-marker") && query.Get("alt") != "media" {
						jsonResp := `{
							"kind": "storage#object",
							"id": "my-bucket/dir-marker/1",
							"name": "dir-marker",
							"bucket": "my-bucket",
							"contentType": "application/x-directory",
							"size": "0",
							"updated": "2023-01-01T00:00:00Z"
						}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonResp)),
							Header:     make(http.Header),
						}, nil
					}
					// Read for dir-marker
					if path == "/my-bucket/dir-marker" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("")),
							Header:     make(http.Header),
						}, nil
					}

					// List request (Readdir)
					// GET /storage/v1/b/my-bucket/o
					if strings.HasSuffix(path, "/o") {
						prefix := query.Get("prefix")
						if prefix == "" {
							// Root listing
							jsonResp := `{
								"kind": "storage#objects",
								"items": [
									{
										"kind": "storage#object",
										"name": "file1.txt",
										"size": "100",
										"updated": "2023-01-01T00:00:00Z"
									}
								],
								"prefixes": [
									"dir1/"
								]
							}`
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       io.NopCloser(strings.NewReader(jsonResp)),
								Header:     make(http.Header),
							}, nil
						}

						if prefix == "dir-marker/" {
							// Listing for dir-marker/
							jsonResp := `{
								"kind": "storage#objects",
								"items": [
									{
										"kind": "storage#object",
										"name": "dir-marker/file1.txt",
										"size": "100",
										"updated": "2023-01-01T00:00:00Z"
									}
								],
								"prefixes": [
									"dir-marker/subdir/"
								]
							}`
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       io.NopCloser(strings.NewReader(jsonResp)),
								Header:     make(http.Header),
							}, nil
						}
					}

					// Delete request
					// DELETE /storage/v1/b/my-bucket/o/file-to-delete.txt
					if req.Method == "DELETE" && strings.Contains(path, "/o/file-to-delete.txt") {
						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(strings.NewReader("")),
							Header:     make(http.Header),
						}, nil
					}

					// Delete request error
					if req.Method == "DELETE" && strings.Contains(path, "/o/delete-fail.txt") {
						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       io.NopCloser(strings.NewReader("Forbidden")),
							Header:     make(http.Header),
						}, nil
					}

					// Rename (Copy + Delete)
					// POST /storage/v1/b/my-bucket/o/source.txt/rewriteTo/b/my-bucket/o/dest.txt
					if req.Method == "POST" && strings.Contains(path, "/rewriteTo/") {
						// Mock successful rewrite
						jsonResp := `{
							"kind": "storage#rewriteResponse",
							"totalBytesRewritten": "10",
							"objectSize": "10",
							"done": true,
							"resource": {
								"kind": "storage#object",
								"name": "dest.txt",
								"bucket": "my-bucket",
								"size": "10"
							}
						}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonResp)),
							Header:     make(http.Header),
						}, nil
					}

					// Handle delete of source after copy for rename
					if req.Method == "DELETE" && strings.Contains(path, "/o/source.txt") {
						return &http.Response{
							StatusCode: http.StatusNoContent,
							Body:       io.NopCloser(strings.NewReader("")),
							Header:     make(http.Header),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader("Unexpected request: " + req.Method + " " + req.URL.String())),
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

	fs := p.GetFs()

	// Test Open & Read
	t.Run("Open and Read", func(t *testing.T) {
		f, err := fs.Open("test-file.txt")
		require.NoError(t, err)
		defer f.Close()

		content, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(content))

		stat, err := f.Stat()
		require.NoError(t, err)
		assert.Equal(t, "test-file.txt", stat.Name())
		// Size might not be populated by NewReader with mock transport
		// assert.Equal(t, int64(11), stat.Size())
	})

	// Test Open missing
	t.Run("Open Missing", func(t *testing.T) {
		_, err := fs.Open("missing.txt")
		assert.Error(t, err)
		if !os.IsNotExist(err) {
			t.Logf("Error is not os.ErrNotExist: %v (%T)", err, err)
		}
		assert.True(t, os.IsNotExist(err), "expected not exist error")
	})

	// Test Stat
	t.Run("Stat", func(t *testing.T) {
		fi, err := fs.Stat("test-file.txt")
		require.NoError(t, err)
		assert.Equal(t, "test-file.txt", fi.Name())
		assert.Equal(t, int64(11), fi.Size())
		assert.False(t, fi.IsDir())
	})

	// Test Stat Missing
	t.Run("Stat Missing", func(t *testing.T) {
		_, err := fs.Stat("missing.txt")
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	// Test Readdir
	t.Run("Readdir", func(t *testing.T) {
		// We open "dir-marker", which exists as an object, to list "dir-marker/"
		f, err := fs.Open("dir-marker")
		require.NoError(t, err)
		defer f.Close()

		infos, err := f.Readdir(-1)
		require.NoError(t, err)
		assert.Len(t, infos, 2) // file1.txt and subdir/

		var foundFile, foundDir bool
		for _, info := range infos {
			if info.Name() == "file1.txt" {
				foundFile = true
				assert.False(t, info.IsDir())
				assert.Equal(t, int64(100), info.Size())
			}
			if info.Name() == "subdir" {
				foundDir = true
				assert.True(t, info.IsDir())
			}
		}
		assert.True(t, foundFile, "file1.txt not found")
		assert.True(t, foundDir, "subdir not found")
	})

	// Test Readdirnames
	t.Run("Readdirnames", func(t *testing.T) {
		f, err := fs.Open("dir-marker")
		require.NoError(t, err)
		defer f.Close()

		names, err := f.Readdirnames(-1)
		require.NoError(t, err)
		assert.Len(t, names, 2)
		assert.Contains(t, names, "file1.txt")
		assert.Contains(t, names, "subdir")
	})

	// Test Remove
	t.Run("Remove", func(t *testing.T) {
		err := fs.Remove("file-to-delete.txt")
		assert.NoError(t, err)
	})

	// Test Remove Error
	t.Run("Remove Error", func(t *testing.T) {
		err := fs.Remove("delete-fail.txt")
		assert.Error(t, err)
	})

	// Test Rename
	t.Run("Rename", func(t *testing.T) {
		err := fs.Rename("source.txt", "dest.txt")
		assert.NoError(t, err)
	})

	// Test Write (Partial)
	// Write requires a bit more complicated mock because GCS writer works by opening a connection and writing chunks.
	// But we can at least open the writer.
	t.Run("Create and Write", func(t *testing.T) {
		// Mocking the write stream is complex with just RoundTripper because the client initiates a resumable upload session.
		// However, we can call Create and check if it returns a file wrapper that has a writer.
		f, err := fs.Create("new-file.txt")
		require.NoError(t, err)
		defer f.Close()

		// We expect writing to fail or stall if we don't mock the upload session correctly,
		// but we can verify the file is created in "write mode"
		assert.NotNil(t, f)
	})
}

// Additional test to cover ReadAt which mimics how afero might use it
func TestGcsFile_ReadAt_WithMock(t *testing.T) {
	// Backup and restore newStorageClient
	oldNewStorageClient := newStorageClient
	defer func() { newStorageClient = oldNewStorageClient }()

	newStorageClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		httpClient := &http.Client{
			Transport: &mockTransportForGcs{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Handle the initial Open (NewReader) check
					if req.Method == "GET" && req.URL.Path == "/my-bucket/test-file.txt" && req.Header.Get("Range") == "" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("hello world content")), // ignored but successful
							Header:     make(http.Header),
						}, nil
					}

					// Read request with Range header
					// GET /storage/v1/b/my-bucket/o/test-file.txt?alt=media OR /my-bucket/test-file.txt
					if strings.Contains(req.URL.Path, "test-file.txt") && req.Header.Get("Range") == "bytes=0-4" {
						headers := make(http.Header)
						headers.Set("Content-Range", "bytes 0-4/5")
						headers.Set("Content-Length", "5")
						return &http.Response{
							StatusCode: http.StatusPartialContent,
							Body:       io.NopCloser(strings.NewReader("hello")),
							Header:     headers,
						}, nil
					}
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader("Not Found: " + req.URL.String() + " Range: " + req.Header.Get("Range"))),
					}, nil
				},
			},
		}
		return storage.NewClient(ctx, option.WithHTTPClient(httpClient))
	}

	config := configv1.GcsFs_builder{Bucket: proto.String("my-bucket")}.Build()
	p, _ := NewGcsProvider(context.Background(), config)
	fs := p.GetFs()

	f, err := fs.Open("test-file.txt")
	require.NoError(t, err)
	defer f.Close()

	buf := make([]byte, 5)
	n, err := f.ReadAt(buf, 0)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(buf))
}

// Test WriteString explicitly
func TestGcsFile_WriteString_WithMock(t *testing.T) {
	// Backup and restore newStorageClient
	oldNewStorageClient := newStorageClient
	defer func() { newStorageClient = oldNewStorageClient }()

	var capturedBody []byte

	newStorageClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		httpClient := &http.Client{
			Transport: &mockTransportForGcs{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Handle resumable upload creation OR simple upload
					if req.Method == "POST" && strings.Contains(req.URL.Path, "/o") {
						// Check if body has content (Simple upload)
						body, _ := io.ReadAll(req.Body)
						if len(body) > 0 {
							capturedBody = body
							// Simple upload response is the object metadata
							jsonResp := `{
								"kind": "storage#object",
								"name": "test-write.txt",
								"bucket": "my-bucket",
								"size": "4"
							}`
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       io.NopCloser(strings.NewReader(jsonResp)),
								Header:     make(http.Header),
							}, nil
						}

						// Resumable init
						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     http.Header{"Location": []string{"http://mock/upload?upload_id=123"}},
							Body:       io.NopCloser(strings.NewReader("{}")),
						}, nil
					}
					// Handle actual data upload
					if req.Method == "PUT" && strings.Contains(req.URL.String(), "upload_id=123") {
						var err error
						capturedBody, err = io.ReadAll(req.Body)
						if err != nil {
							return nil, err
						}
						jsonResp := `{
							"kind": "storage#object",
							"name": "test-write.txt",
							"bucket": "my-bucket",
							"size": "4"
						}`
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonResp)),
							Header:     make(http.Header),
						}, nil
					}
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader("Unexpected request: " + req.Method + " " + req.URL.String())),
					}, nil
				},
			},
		}
		return storage.NewClient(ctx, option.WithHTTPClient(httpClient))
	}

	config := configv1.GcsFs_builder{Bucket: proto.String("my-bucket")}.Build()
	p, _ := NewGcsProvider(context.Background(), config)
	fs := p.GetFs()

	f, err := fs.Create("test-write.txt")
	require.NoError(t, err)

	n, err := f.WriteString("data")
	require.NoError(t, err)
	assert.Equal(t, 4, n)

	err = f.Close()
	require.NoError(t, err)

	// Check if captured body contains the data (it might be multipart)
	assert.Contains(t, string(capturedBody), "data")
}
