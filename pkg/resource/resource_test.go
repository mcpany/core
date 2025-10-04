/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResource is a mock implementation of the Resource interface for testing.
type mockResource struct {
	uri          string
	service      string
	subscribeErr error
}

func (r *mockResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: r.uri}
}

func (r *mockResource) Service() string {
	return r.service
}

func (r *mockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{}, nil
}

func (r *mockResource) Subscribe(ctx context.Context) error {
	return r.subscribeErr
}

func TestNewResourceManager(t *testing.T) {
	rm := NewResourceManager()
	assert.NotNil(t, rm)
	assert.NotNil(t, rm.resources)
}

func TestResourceManager_AddGetListRemoveResource(t *testing.T) {
	rm := NewResourceManager()
	resource1 := &mockResource{uri: "resource://one", service: "service1"}
	resource2 := &mockResource{uri: "resource://two", service: "service2"}

	// Add
	rm.AddResource(resource1)
	rm.AddResource(resource2)

	// Get
	r, ok := rm.GetResource("resource://one")
	require.True(t, ok)
	assert.Equal(t, resource1, r)

	r, ok = rm.GetResource("resource://two")
	require.True(t, ok)
	assert.Equal(t, resource2, r)

	_, ok = rm.GetResource("non-existent")
	assert.False(t, ok)

	// List
	resources := rm.ListResources()
	assert.Len(t, resources, 2)
	assert.Contains(t, resources, resource1)
	assert.Contains(t, resources, resource2)

	// Remove
	rm.RemoveResource("resource://one")
	_, ok = rm.GetResource("resource://one")
	assert.False(t, ok)
	assert.Len(t, rm.ListResources(), 1)
}

func TestResourceManager_OnListChanged(t *testing.T) {
	rm := NewResourceManager()
	var changedCount int
	rm.OnListChanged(func() {
		changedCount++
	})

	// Add should trigger the callback
	rm.AddResource(&mockResource{uri: "r1"})
	assert.Equal(t, 1, changedCount, "OnListChanged should be called on AddResource")

	// Remove should trigger the callback
	rm.RemoveResource("r1")
	assert.Equal(t, 2, changedCount, "OnListChanged should be called on RemoveResource")

	// Removing a non-existent resource should not trigger the callback
	rm.RemoveResource("non-existent")
	assert.Equal(t, 2, changedCount, "OnListChanged should not be called for non-existent resource removal")
}

func TestResourceManager_Subscribe(t *testing.T) {
	rm := NewResourceManager()

	t.Run("subscribe success", func(t *testing.T) {
		res := &mockResource{uri: "res1"}
		rm.AddResource(res)
		err := rm.Subscribe(context.Background(), "res1")
		assert.NoError(t, err)
	})

	t.Run("subscribe not found", func(t *testing.T) {
		err := rm.Subscribe(context.Background(), "not-found")
		assert.Error(t, err)
		assert.Equal(t, ErrResourceNotFound, err)
	})

	t.Run("subscribe error", func(t *testing.T) {
		subscribeErr := errors.New("subscribe failed")
		res := &mockResource{uri: "res2", subscribeErr: subscribeErr}
		rm.AddResource(res)
		err := rm.Subscribe(context.Background(), "res2")
		assert.Error(t, err)
		assert.Equal(t, subscribeErr, err)
	})
}
