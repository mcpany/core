// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package util provides utility functions for Docker and other shared functionality.
package util //nolint:revive,nolintlint // Package name 'util' is intentional

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
)

var (
	// IsDockerSocketAccessibleFunc is a variable to allow mocking in tests.
	// It checks if the Docker socket is accessible.
	// Summary: Defines IsDockerSocketAccessibleFunc.
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault

	dockerClient     client.APIClient
	initDockerClient = initDockerClientDefault
	once             = &sync.Once{}
)

// initDockerClientDefault initializes the shared Docker client. This function is
// intended to be called only once.
var initDockerClientDefault = func() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// If we can't create the client, we can't ping the server.
		// We'll set dockerClient to nil and handle this in the check.
		dockerClient = nil
	}
}

// IsDockerSocketAccessible serves as a public interface for interacting with IsDockerSocketAccessible.
//
// Summary: Checks condition indicating whether the target is docker socket accessible.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func IsDockerSocketAccessible() bool {
	return IsDockerSocketAccessibleFunc()
}

// CloseDockerClient serves as a public interface for interacting with CloseDockerClient.
//
// Summary: Close the docker client appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func CloseDockerClient() {
	if dockerClient != nil {
		_ = dockerClient.Close()
	}
}

// isDockerSocketAccessibleDefault is the default implementation for checking
// Docker socket accessibility. It pings the Docker daemon to verify that it is
// running and accessible.
func isDockerSocketAccessibleDefault() bool {
	once.Do(initDockerClient)

	if dockerClient == nil {
		return false
	}

	_, err := dockerClient.Ping(context.Background())
	return err == nil
}
