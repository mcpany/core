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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDockerSocketAccessible(t *testing.T) {
	// Reset state before test
	dockerSocketCheckMutex.Lock()
	dockerSocketCheckCompleted = false
	dockerSocketCheckMutex.Unlock()

	// Call the function for the first time. This will perform the actual check.
	accessible := IsDockerSocketAccessible()

	// We can't guarantee the Docker socket is available in the test environment,
	// so we just check that it returns a boolean value without panicking.
	// In a CI environment with Docker, this should be true. Without, it will be false.
	assert.NotPanics(t, func() { IsDockerSocketAccessible() })

	// Call it a second time to ensure the cached result is returned.
	cachedResult := IsDockerSocketAccessible()
	assert.Equal(t, accessible, cachedResult, "Cached result should be the same as the first call")

	// Verify that the check was marked as completed.
	dockerSocketCheckMutex.Lock()
	assert.True(t, dockerSocketCheckCompleted, "Check should be marked as completed")
	dockerSocketCheckMutex.Unlock()
}
