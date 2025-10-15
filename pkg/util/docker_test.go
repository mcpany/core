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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDockerSocketAccessible(t *testing.T) {
	// We can't guarantee the Docker socket is available in the test environment,
	// so we just check that it returns a boolean value without panicking.
	// In a CI environment with Docker, this should be true. Without, it will be false.
	assert.NotPanics(t, func() { IsDockerSocketAccessible() })
}

func TestIsDockerSocketAccessible_StateChange(t *testing.T) {
	// Restore the original function after the test
	defer func() { IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault }()

	// Mock the function to return true
	IsDockerSocketAccessibleFunc = func() bool { return true }
	assert.True(t, IsDockerSocketAccessible(), "Should return true when Docker is accessible")

	// Mock the function to return false
	IsDockerSocketAccessibleFunc = func() bool { return false }
	assert.False(t, IsDockerSocketAccessible(), "Should return false when Docker is not accessible")
}
