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
		config := &configv1.S3Fs{
			Bucket:          proto.String("my-bucket"),
			Region:          proto.String("us-east-1"),
			AccessKeyId:     proto.String("test"),
			SecretAccessKey: proto.String("test"),
			Endpoint:        proto.String("http://localhost:9000"),
		}
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.GetFs())
		p.Close()
	})

	t.Run("Missing Bucket", func(t *testing.T) {
		// afero-s3 doesn't error on missing bucket at creation time usually
		config := &configv1.S3Fs{
			Region: proto.String("us-east-1"),
		}
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Region Only", func(t *testing.T) {
		config := &configv1.S3Fs{
			Region: proto.String("us-west-2"),
		}
		p, err := NewS3Provider(config)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}
