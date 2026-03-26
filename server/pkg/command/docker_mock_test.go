// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// MockDockerClient ...
// Summary: MockDockerClient
	ImagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFunc  func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogsFunc   func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerWaitFunc   func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerAttachFunc func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
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
	return io.NopCloser(strings.NewReader("")), nil
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
	return container.CreateResponse{ID: "test-container-id"}, nil
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
		return m.ContainerStartFunc(ctx, containerID, options)
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

// ContainerLogs ...
// Summary: ContainerLogs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerLogsFunc != nil {
		return m.ContainerLogsFunc(ctx, container, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

// ContainerWait ...
// Summary: ContainerWait
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ContainerWaitFunc != nil {
		return m.ContainerWaitFunc(ctx, containerID, condition)
	}
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	statusCh <- container.WaitResponse{StatusCode: 0}
	close(statusCh)
	close(errCh)
	return statusCh, errCh
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
