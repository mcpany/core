// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewS3Provider(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		config := configv1.S3Fs_builder{
			Bucket:          proto.String("my-bucket"),
			Region:          proto.String("us-east-1"),
			AccessKeyId:     proto.String("test"),
			SecretAccessKey: proto.String("test"),
			Endpoint:        proto.String("http://127.0.0.1:9000"),
		}.Build()
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.GetFs())
		p.Close()
	})

	t.Run("Missing Bucket", func(t *testing.T) {
		// afero-s3 doesn't error on missing bucket at creation time usually
		config := configv1.S3Fs_builder{
			Region: proto.String("us-east-1"),
		}.Build()
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Region Only", func(t *testing.T) {
		config := configv1.S3Fs_builder{
			Region: proto.String("us-west-2"),
		}.Build()
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}

func TestS3Provider_ResolvePath(t *testing.T) {
	config := configv1.S3Fs_builder{
		Bucket: proto.String("my-bucket"),
		Region: proto.String("us-east-1"),
	}.Build()
	p, err := NewS3Provider(config)
	require.NoError(t, err)
	defer p.Close()

	tests := []struct {
		name        string
		virtualPath string
		want        string
		wantErr     bool
	}{
		{
			name:        "Root",
			virtualPath: "/",
			want:        "", // TrimPrefix removes leading slash, path.Clean("/") is "/"
			// Wait, if cleanPath is "", it returns error "invalid path"
			wantErr: true,
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
			name:        "No leading slash",
			virtualPath: "file.txt",
			want:        "file.txt",
			wantErr:     false,
		},
		{
			name:        "Traversal",
			virtualPath: "../file.txt",
			want:        "file.txt", // Sandboxed to root
			wantErr:     false,
		},
		{
			name:        "Deep traversal",
			virtualPath: "a/../../b",
			want:        "b",
			wantErr:     false,
		},
		{
			name:        "Dot",
			virtualPath: ".",
			wantErr:     true,
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
