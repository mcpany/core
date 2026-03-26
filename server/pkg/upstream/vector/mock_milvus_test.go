// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// mockMilvusClient implements client.Client interface for testing.
// Only methods used by MilvusClient are implemented.
type mockMilvusClient struct {
	client.Client // Embed to satisfy interface, will panic if unimplemented methods are called

	// Mock function implementations
	hasCollectionFunc           func(ctx context.Context, name string) (bool, error)
	loadCollectionFunc          func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error
	describeCollectionFunc      func(ctx context.Context, name string) (*entity.Collection, error)
	searchFunc                  func(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam, opts ...client.SearchQueryOptionFunc) ([]client.SearchResult, error)
	upsertFunc                  func(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) (entity.Column, error)
	deleteFunc                  func(ctx context.Context, collectionName string, partitionName string, expr string) error
	getCollectionStatisticsFunc func(ctx context.Context, name string) (map[string]string, error)
	closeFunc                   func() error
}

// HasCollection ...
// Summary: HasCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.hasCollectionFunc != nil {
		return m.hasCollectionFunc(ctx, name)
	}
	return false, nil
}

// LoadCollection ...
// Summary: LoadCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.loadCollectionFunc != nil {
		return m.loadCollectionFunc(ctx, name, async, opts...)
	}
	return nil
}

// DescribeCollection ...
// Summary: DescribeCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.describeCollectionFunc != nil {
		return m.describeCollectionFunc(ctx, name)
	}
	return nil, nil
}

// Search ...
// Summary: Search
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.searchFunc != nil {
		return m.searchFunc(ctx, collectionName, partitions, expr, outputFields, vectors, vectorField, metricType, topK, sp, opts...)
	}
	return nil, nil
}

// Upsert ...
// Summary: Upsert
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, collectionName, partitionName, columns...)
	}
	return nil, nil
}

// Delete ...
// Summary: Delete
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, collectionName, partitionName, expr)
	}
	return nil
}

// GetCollectionStatistics ...
// Summary: GetCollectionStatistics
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.getCollectionStatisticsFunc != nil {
		return m.getCollectionStatisticsFunc(ctx, name)
	}
	return nil, nil
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
