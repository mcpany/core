// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigWatcher(t *testing.T) {
	dir, err := os.MkdirTemp("", "config_watcher_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "config.yaml")
	err = os.WriteFile(file, []byte("foo"), 0644)
	require.NoError(t, err)

	reloaded := make(chan struct{})
	reload := func() error {
		close(reloaded)
		return nil
	}

	cw, err := NewConfigWatcher([]string{file}, reload)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cw.Start(ctx)
	defer cw.Stop()

	err = os.WriteFile(file, []byte("bar"), 0644)
	require.NoError(t, err)

	select {
	case <-reloaded:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for reload")
	}
}
