// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestDefaultClientFactory_UnsupportedType(t *testing.T) {
	config := &configv1.VectorUpstreamService{
		VectorDbType: nil, // Unsupported
	}
	client, err := defaultClientFactory(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unsupported vector database type")
}

func TestDefaultClientFactory_Milvus(t *testing.T) {
	// This checks that the factory correctly routes to NewMilvusClient
	// even if connection fails.
	config := &configv1.VectorUpstreamService{
		VectorDbType: &configv1.VectorUpstreamService_Milvus{
			Milvus: &configv1.MilvusVectorDB{
				Address:        proto.String("127.0.0.1:19530"),
				CollectionName: proto.String("test"),
			},
		},
	}
	client, err := defaultClientFactory(config)
	// We expect an error because we don't have a running Milvus instance,
	// but this confirms we hit the Milvus case and attempted creation.
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestUpstream_Register_SanitizeError(t *testing.T) {
	u := NewUpstream()
	// SanitizeServiceName returns error for empty name
	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String(""),
	}

	_, _, _, err := u.Register(context.Background(), cfg, nil, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")
}
