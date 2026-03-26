// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestS3Provider_Mock(t *testing.T) {
	// Start a mock S3 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock S3 received: %s %s", r.Method, r.URL.Path)

		// Verify bucket in path (ForcePathStyle=true)
		if !strings.HasPrefix(r.URL.Path, "/my-mock-bucket") {
			t.Logf("Bucket not found in path: %s", r.URL.Path)
			http.Error(w, "Bucket not found", http.StatusNotFound)
			return
		}

		// Handle implicit root or explicit root
		key := strings.TrimPrefix(r.URL.Path, "/my-mock-bucket/")
		// If path was exactly /my-mock-bucket, TrimPrefix returns it unchanged if it doesn't end in /
		if r.URL.Path == "/my-mock-bucket" {
			key = ""
		}

		if key == "" || key == "/" {
			// List Objects (ListBucket)
			w.Header().Set("Content-Type", "application/xml")
			// Minimal XML response for ListObjectsV2
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
    <Name>my-mock-bucket</Name>
    <Prefix></Prefix>
    <KeyCount>1</KeyCount>
    <MaxKeys>1000</MaxKeys>
    <IsTruncated>false</IsTruncated>
    <Contents>
        <Key>hello.txt</Key>
        <LastModified>2023-01-01T00:00:00.000Z</LastModified>
        <ETag>"d41d8cd98f00b204e9800998ecf8427e"</ETag>
        <Size>11</Size>
        <StorageClass>STANDARD</StorageClass>
    </Contents>
</ListBucketResult>`)
			return
		} else if key == "hello.txt" {
			// Get Object
			if r.Method == "HEAD" {
				w.Header().Set("Content-Length", "11")
				w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
				w.Header().Set("ETag", "\"d41d8cd98f00b204e9800998ecf8427e\"")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Content-Length", "11")
			w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			w.Header().Set("ETag", "\"d41d8cd98f00b204e9800998ecf8427e\"")
			fmt.Fprint(w, "Hello World")
			return
		}

		http.Error(w, "Key not found", http.StatusNotFound)
	}))
	defer server.Close()

	// Configure S3Provider to use the mock server
	config := configv1.S3Fs_builder{
		Bucket:          proto.String("my-mock-bucket"),
		Region:          proto.String("us-east-1"),
		AccessKeyId:     proto.String("test"),
		SecretAccessKey: proto.String("test"),
		Endpoint:        proto.String(server.URL),
	}.Build()

	p, err := NewS3Provider(config)
	require.NoError(t, err)
	defer p.Close()

	fs := p.GetFs()

	// Test 1: List Directory
	// Note: afero-s3 implementation of Open("/") usually triggers a ListObjects
	f, err := fs.Open("/")
	require.NoError(t, err)
	defer f.Close()
	infos, err := f.Readdir(-1)
	require.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, "hello.txt", infos[0].Name())

	// Test 2: Read File
	// Open the file explicitly
	file, err := fs.Open("hello.txt")
	require.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", stat.Name())
	assert.Equal(t, int64(11), stat.Size())

	// Read content
	buf := make([]byte, 11)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		require.NoError(t, err)
	}
	assert.Equal(t, 11, n)
	assert.Equal(t, "Hello World", string(buf))
}
