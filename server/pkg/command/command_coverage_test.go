// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWriteOnlyConn implements net.Conn for the Docker attach stdin side only.
// Writing goes to a pipe writer; CloseWrite/Close closes the write end of that
// pipe, signalling EOF to the simulated container without affecting the reader.
type mockWriteOnlyConn struct {
	w *io.PipeWriter
}

func (c *mockWriteOnlyConn) Write(p []byte) (int, error)        { return c.w.Write(p) }
func (c *mockWriteOnlyConn) CloseWrite() error                  { return c.w.Close() }
func (c *mockWriteOnlyConn) Close() error                       { return c.w.Close() }
func (c *mockWriteOnlyConn) Read(_ []byte) (int, error)         { return 0, io.EOF }
func (c *mockWriteOnlyConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *mockWriteOnlyConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *mockWriteOnlyConn) SetDeadline(_ time.Time) error      { return nil }
func (c *mockWriteOnlyConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *mockWriteOnlyConn) SetWriteDeadline(_ time.Time) error { return nil }

// TestDockerExecutorWithStdIO_Mock exercises the full stdin→stdout round-trip
// of dockerExecutor.ExecuteWithStdIO without requiring a real Docker daemon.
// It is the canonical replacement for the real-Docker "Success" subtest that
// is skipped due to DinD unreliability in CI.
func TestDockerExecutorWithStdIO_Mock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// stdinPR/stdinPW: written by the executor (via the closeWriter), read
		// by the fake container goroutine below.
		stdinPR, stdinPW := io.Pipe()

		// serverPR/serverPW: the fake container writes multiplexed stdout here;
		// the executor's stdcopy goroutine reads from serverPR.
		serverPR, serverPW := io.Pipe()

		// Container-exit channels – buffered so the fake goroutine never blocks.
		statusCh := make(chan container.WaitResponse, 1)
		errCh := make(chan error, 1)

		// Fake container: echo stdin as multiplexed stdout, then signal exit.
		go func() {
			defer serverPW.Close()
			data, _ := io.ReadAll(stdinPR)

			// Docker multiplexed-stream frame: 1 byte stream type (1=stdout),
			// 3 zero bytes, 4 bytes big-endian payload length.
			header := make([]byte, 8)
			header[0] = 1 // stdout
			binary.BigEndian.PutUint32(header[4:], uint32(len(data)))
			_, _ = serverPW.Write(header)
			_, _ = serverPW.Write(data)

			statusCh <- container.WaitResponse{StatusCode: 0}
			close(statusCh)
			close(errCh)
		}()

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{
					Conn:   &mockWriteOnlyConn{w: stdinPW},
					Reader: bufio.NewReader(serverPR),
				}, nil
			},
			ContainerWaitFunc: func(_ context.Context, _ string, _ container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
				return statusCh, errCh
			},
		}
		executor.clientFactory = func() (DockerClient, error) { return mockClient, nil }

		stdin, stdout, stderr, exitCodeChan, err := executor.ExecuteWithStdIO(
			context.Background(), "cat", nil, "", nil)
		require.NoError(t, err)

		_, err = stdin.Write([]byte("hello\n"))
		require.NoError(t, err)
		require.NoError(t, stdin.Close())

		stdoutData, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello\n", string(stdoutData))

		stderrData, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, stderrData)

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})
}

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
