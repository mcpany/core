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

	"github.com/docker/docker/client"
)

var (
	// IsDockerSocketAccessibleFunc is a function that can be replaced for testing purposes.
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault
)

// IsDockerSocketAccessible checks if the Docker daemon is accessible through the socket.
func IsDockerSocketAccessible() bool {
	return IsDockerSocketAccessibleFunc()
}

func isDockerSocketAccessibleDefault() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer cli.Close()

	_, err = cli.Ping(context.Background())
	return err == nil
}
