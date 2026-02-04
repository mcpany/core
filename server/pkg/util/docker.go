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
	//
	// Summary: Checks if the Docker socket is accessible.
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

// IsDockerSocketAccessible checks if the Docker daemon is accessible through the socket.
//
// Summary: Checks if the Docker daemon is accessible through the socket.
//
// Returns:
//   - bool: True if the Docker daemon is accessible, false otherwise.
func IsDockerSocketAccessible() bool {
	return IsDockerSocketAccessibleFunc()
}

// CloseDockerClient closes the shared Docker client.
//
// Summary: Closes the shared Docker client.
//
// Side Effects:
//   - Closes the connection to the Docker daemon if it was established.
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
