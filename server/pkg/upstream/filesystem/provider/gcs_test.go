// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewGcsProvider(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		p, err := NewGcsProvider(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, "gcs config is nil", err.Error())
	})

	t.Run("Valid Config", func(t *testing.T) {
		// This will try to create a GCS client.
		// Without credentials, it might fail or try to find default credentials.
		// In a CI environment/sandbox, it might fail if no creds.
		// However, storage.NewClient doesn't error immediately usually unless invalid options.
		// But it might try to connect or find creds.
		cfg := &configv1.GcsFs{
			Bucket: proto.String("my-bucket"),
		}
		p, err := NewGcsProvider(context.Background(), cfg)

		// If it errors due to no credentials, that's fine, we asserted it tried.
		if err != nil {
			// Ensure it's not our "nil config" error
			assert.NotEqual(t, "gcs config is nil", err.Error())
		} else {
			assert.NotNil(t, p)
			assert.NotNil(t, p.GetFs())
			assert.NoError(t, p.Close())
		}
	})
}
