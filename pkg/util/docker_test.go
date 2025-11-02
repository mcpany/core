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

package util

import (
	"sync"
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
	defer func() {
		IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault
		// Reset the singleton for other tests
		once = sync.Once{}
		if dockerClient != nil {
			dockerClient.Close()
			dockerClient = nil
		}
	}()

	// Mock the function to return true
	IsDockerSocketAccessibleFunc = func() bool { return true }
	assert.True(t, IsDockerSocketAccessible(), "Should return true when Docker is accessible")

	// Mock the function to return false
	IsDockerSocketAccessibleFunc = func() bool { return false }
	assert.False(t, IsDockerSocketAccessible(), "Should return false when Docker is not accessible")
}

func TestDockerClient_Singleton(t *testing.T) {
	// Restore the original init function and reset the singleton after the test
	originalInit := initDockerClient
	defer func() {
		initDockerClient = originalInit
		once = sync.Once{}
		if dockerClient != nil {
			dockerClient.Close()
			dockerClient = nil
		}
	}()

	var initializationCount int
	// Replace the init function with a mock that counts calls
	initDockerClient = func() {
		initializationCount++
		originalInit() // Call the original init function to maintain behavior
	}

	// Call the function multiple times
	isDockerSocketAccessibleDefault()
	isDockerSocketAccessibleDefault()

	// Check that the initialization function was called only once
	assert.Equal(t, 1, initializationCount, "The Docker client should be initialized only once")
}

func TestCloseDockerClient(t *testing.T) {
	// Restore the original init function and reset the singleton after the test
	originalInit := initDockerClient
	defer func() {
		initDockerClient = originalInit
		once = sync.Once{}
		if dockerClient != nil {
			dockerClient.Close()
			dockerClient = nil
		}
	}()

	// Ensure the client is initialized
	isDockerSocketAccessibleDefault()
	assert.NotNil(t, dockerClient, "Docker client should be initialized")

	// Close the client
	CloseDockerClient()
}

func TestInitDockerClient_Error(t *testing.T) {
	// Restore the original init function and reset the singleton after the test
	originalInit := initDockerClient
	defer func() {
		initDockerClient = originalInit
		once = sync.Once{}
		if dockerClient != nil {
			dockerClient.Close()
			dockerClient = nil
		}
	}()

	// Replace the init function with a mock that always fails
	initDockerClient = func() {
		dockerClient = nil
	}

	// Call the function that uses the client
	accessible := isDockerSocketAccessibleDefault()

	// Check that the client is nil and the function returns false
	assert.Nil(t, dockerClient, "Docker client should be nil after a failed initialization")
	assert.False(t, accessible, "isDockerSocketAccessibleDefault should return false when initialization fails")
}
