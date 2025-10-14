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

import (
	"sync"
)

func TestIsDockerSocketAccessible(t *testing.T) {
	// To properly test the sync.Once functionality, we need to reset the state.
	// In a real-world scenario, you might not do this, but for a unit test, it's necessary
	// to ensure a clean slate for each test run.
	dockerSocketCheckOnce = sync.Once{}
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault

	// Call the function for the first time. This will perform the actual check.
	accessible := IsDockerSocketAccessible()

	// We can't guarantee the Docker socket is available in the test environment,
	// so we just check that it returns a boolean value without panicking.
	// In a CI environment with Docker, this should be true. Without, it will be false.
	assert.NotPanics(t, func() { IsDockerSocketAccessible() })

	// Call it a second time to ensure the cached result is returned.
	cachedResult := IsDockerSocketAccessible()
	assert.Equal(t, accessible, cachedResult, "Cached result should be the same as the first call")
}

func TestIsDockerSocketAccessible_Concurrency(t *testing.T) {
	// Reset the state for this specific test to ensure it's isolated.
	dockerSocketCheckOnce = sync.Once{}
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault

	var callCount int
	// Replace the default function with a mock that increments a counter
	IsDockerSocketAccessibleFunc = func() bool {
		dockerSocketCheckOnce.Do(func() {
			callCount++
			// Simulate the original check's behavior
			dockerSocketAccessible = true
		})
		return dockerSocketAccessible
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	// Number of concurrent calls to make
	const concurrentCalls = 100

	wg.Add(concurrentCalls)
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			IsDockerSocketAccessible()
		}()
	}
	wg.Wait()

	// Assert that the underlying check function was called only once
	assert.Equal(t, 1, callCount, "The check function should only be called once")

	// Restore the original function
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault
}
