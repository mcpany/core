package command

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerExecutorWithStdIO_ErrorPaths(t *testing.T) {
	t.Run("ExecuteWithStdIO_ContainerAttachError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			return types.HijackedResponse{}, errors.New("attach error")
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attach error")
	})

	t.Run("ExecuteWithStdIO_ContainerStartError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerStartFunc = func(ctx context.Context, containerID string, options container.StartOptions) error {
			return errors.New("start error")
		}
		mockClient.ContainerRemoveFunc = func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return nil
		}

		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			s, c := net.Pipe()
			defer s.Close()
			return types.HijackedResponse{
				Conn:   c,
				Reader: bufio.NewReader(c),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start error")
	})

	t.Run("ExecuteWithStdIO_ContainerStartError_RemoveFails", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerStartFunc = func(ctx context.Context, containerID string, options container.StartOptions) error {
			return errors.New("start error")
		}
		mockClient.ContainerRemoveFunc = func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return errors.New("remove error")
		}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			s, c := net.Pipe()
			defer s.Close()
			return types.HijackedResponse{
				Conn:   c,
				Reader: bufio.NewReader(c),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start error")
	})

	t.Run("ExecuteWithStdIO_ContainerWaitError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerWaitFunc = func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
			errCh := make(chan error, 1)
			errCh <- errors.New("wait error")
			return nil, errCh
		}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			s, c := net.Pipe()
			defer s.Close()
			return types.HijackedResponse{
				Conn:   c,
				Reader: bufio.NewReader(c),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, -1, exitCode)
	})

	t.Run("ExecuteWithStdIO_ClientFactoryError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		executor.clientFactory = func() (DockerClient, error) {
			return nil, errors.New("client factory error")
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client factory error")
	})

	t.Run("ExecuteWithStdIO_ImagePullError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ImagePullFunc = func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
			return nil, errors.New("pull error")
		}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			s, c := net.Pipe()
			defer s.Close()
			return types.HijackedResponse{
				Conn:   c,
				Reader: bufio.NewReader(c),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("ExecuteWithStdIO_Local_CommandNotFound", func(t *testing.T) {
		executor := NewLocalExecutor()
		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "non-existent-command", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "executable file not found")
	})
}
