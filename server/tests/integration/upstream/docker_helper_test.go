// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// canConnectToDocker checks if we can connect to Docker AND run a container.
// This is critical for CI environments (DinD) where the daemon might be reachable
// but unable to mount overlays/volumes due to storage driver issues.
func canConnectToDocker(t *testing.T) bool {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		return false
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Logf("could not create docker client: %v", err)
		return false
	}
	defer cli.Close()

	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Logf("could not ping docker daemon: %v", err)
		return false
	}

	// Verify we can actually run a container.
	// This catches issues like "mount source overlay ... invalid argument" in DinD environments.
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"true"},
	}, nil, nil, nil, "")
	if err != nil {
		t.Logf("Docker daemon reachable but failed to create container (skipping tests): %v", err)
		return false
	}
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.Logf("Docker daemon reachable but failed to start container (skipping tests): %v", err)
		return false
	}

	return true
}
