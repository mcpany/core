/*
 * Copyright 2025 Author(s) of MCPX
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

package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpPool_New(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		p, err := NewHttpPool(1, 5, 100)
		require.NoError(t, err)
		assert.NotNil(t, p)
		defer p.Close()

		assert.Equal(t, 1, p.Len())

		client, err := p.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy())

		p.Put(client)
		assert.Equal(t, 1, p.Len())
	})

	t.Run("invalid config", func(t *testing.T) {
		_, err := NewHttpPool(5, 1, 10)
		assert.Error(t, err)
	})
}

func TestHttpPool_GetPut(t *testing.T) {
	p, err := NewHttpPool(1, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, p)

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.IsHealthy())

	p.Put(client)

	// After putting it back, we should be able to get it again.
	client2, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client2)

	// It should be the same client since pool size is 1
	assert.Same(t, client, client2)
}

func TestHttpPool_Close(t *testing.T) {
	p, err := NewHttpPool(1, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, p)

	p.Close()

	// After closing, get should fail
	_, err = p.Get(context.Background())
	assert.Error(t, err)
}

func TestHttpPool_PoolFull(t *testing.T) {
	p, err := NewHttpPool(1, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, p)

	// Get the only client
	_, err = p.Get(context.Background())
	require.NoError(t, err)

	// Try to get another one, should time out
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = p.Get(ctx)
	assert.Error(t, err)
}
