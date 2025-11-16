/*
 * Copyright 2025 Author(s) of MCP Any
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

package app

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	app := NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find a free port for the server
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	listener.Close()

	go func() {
		err := app.Run(ctx, afero.NewMemMapFs(), false, addr, "", nil, 5*time.Second)
		assert.NoError(t, err)
	}()

	// Wait for the server to start
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://%s/healthz", addr))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "server did not start in time")

	var out bytes.Buffer
	err = HealthCheck(&out, addr, 5*time.Second)
	assert.NoError(t, err)
	assert.Contains(t, out.String(), "Health check successful")
}
