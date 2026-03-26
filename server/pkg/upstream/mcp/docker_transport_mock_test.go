// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockDockerClient struct {
	ImagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerAttachFunc func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	ContainerStartFunc  func(ctx context.Context, container string, options container.StartOptions) error
	ContainerStopFunc   func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
	CloseFunc           func() error
}

// ImagePull ...
// Summary: ImagePull
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ImagePullFunc != nil {
		return m.ImagePullFunc(ctx, ref, options)
	}
	return nil, nil
}

// ContainerCreate ...
// Summary: ContainerCreate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerCreateFunc != nil {
		return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{}, nil
}

// ContainerAttach ...
// Summary: ContainerAttach
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerAttachFunc != nil {
		return m.ContainerAttachFunc(ctx, container, options)
	}
	return types.HijackedResponse{}, nil
}

// ContainerStart ...
// Summary: ContainerStart
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerStartFunc != nil {
		return m.ContainerStartFunc(ctx, container, options)
	}
	return nil
}

// ContainerStop ...
// Summary: ContainerStop
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerStopFunc != nil {
		return m.ContainerStopFunc(ctx, containerID, options)
	}
	return nil
}

// ContainerRemove ...
// Summary: ContainerRemove
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerRemoveFunc != nil {
		return m.ContainerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
