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

package util

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
)

var (
	dockerSocketAccessible     bool
	dockerSocketCheckCompleted bool
	dockerSocketCheckMutex     sync.Mutex
)

var IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault

// IsDockerSocketAccessible checks if the Docker daemon is accessible through the socket.
// It caches the result to avoid repeated checks.
func IsDockerSocketAccessible() bool {
	return IsDockerSocketAccessibleFunc()
}

func isDockerSocketAccessibleDefault() bool {
	dockerSocketCheckMutex.Lock()
	defer dockerSocketCheckMutex.Unlock()

	if dockerSocketCheckCompleted {
		return dockerSocketAccessible
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		dockerSocketAccessible = false
	} else {
		defer cli.Close()
		_, err = cli.Ping(context.Background())
		dockerSocketAccessible = err == nil
	}

	dockerSocketCheckCompleted = true
	return dockerSocketAccessible
}
