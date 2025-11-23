
package util

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
)

var (
	// IsDockerSocketAccessibleFunc is a function that can be replaced for testing purposes.
	IsDockerSocketAccessibleFunc = isDockerSocketAccessibleDefault

	dockerClient *client.Client
	once         sync.Once
)

// initDockerClient initializes the shared Docker client. This function is
// intended to be called only once.
var initDockerClient = func() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// If we can't create the client, we can't ping the server.
		// We'll set dockerClient to nil and handle this in the check.
		dockerClient = nil
	}
}

// IsDockerSocketAccessible checks if the Docker daemon is accessible through the socket.
func IsDockerSocketAccessible() bool {
	return IsDockerSocketAccessibleFunc()
}

// CloseDockerClient closes the shared Docker client.
// It should be called on application shutdown.
func CloseDockerClient() {
	if dockerClient != nil {
		dockerClient.Close()
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
