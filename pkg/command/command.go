// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package command provides interfaces and implementations for executing commands.
package command

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Executor is an interface for executing commands.
type Executor interface {
	Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
	ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (stdin io.WriteCloser, stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
}

// DockerClient is an interface for Docker client methods used by dockerExecutor.
// This allows mocking the Docker client in tests.
type DockerClient interface {
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	Close() error
}

// newDockerClientFunc allows mocking the docker client in tests.
var newDockerClientFunc = func() (DockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// NewExecutor creates a new command executor.
func NewExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	if containerEnv != nil && containerEnv.GetImage() != "" {
		return newDockerExecutor(containerEnv)
	}
	return &localExecutor{}
}

// NewLocalExecutor creates a new local command executor.
func NewLocalExecutor() Executor {
	return &localExecutor{}
}

type localExecutor struct{}

func (e *localExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDir
	cmd.Env = env

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	cmd.Stdout = outW
	cmd.Stderr = errW

	if err := cmd.Start(); err != nil {
		_ = outW.Close()
		_ = errW.Close()
		return nil, nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer close(exitCodeChan)
		defer func() { _ = outW.Close() }()
		defer func() { _ = errW.Close() }()

		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCodeChan <- exitErr.ExitCode()
			} else {
				logging.GetLogger().Error("Command execution failed", "error", err)
				exitCodeChan <- -1
			}
		} else {
			exitCodeChan <- 0
		}
	}()

	return outR, errR, exitCodeChan, nil
}

func (e *localExecutor) ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDir
	cmd.Env = env

	// Use io.Pipe for proper synchronization and avoidance of race conditions
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	if err := cmd.Start(); err != nil {
		_ = stdinR.Close()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		_ = stderrR.Close()
		_ = stderrW.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer close(exitCodeChan)
		// Close the write ends of the output pipes so the reader gets EOF
		defer func() { _ = stdoutW.Close() }()
		defer func() { _ = stderrW.Close() }()

		// Also ensure stdin read end is closed if not already
		defer func() { _ = stdinR.Close() }()

		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCodeChan <- exitErr.ExitCode()
			} else {
				logging.GetLogger().Error("Command execution failed", "error", err)
				exitCodeChan <- -1
			}
		} else {
			exitCodeChan <- 0
		}
	}()

	return stdinW, stdoutR, stderrR, exitCodeChan, nil
}

type dockerExecutor struct {
	containerEnv *configv1.ContainerEnvironment
}

func newDockerExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	return &dockerExecutor{containerEnv: containerEnv}
}

func (e *dockerExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	log := logging.GetLogger()
	cli, err := newDockerClientFunc()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	img := e.containerEnv.GetImage()
	reader, err := cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		log.Warn("Failed to pull docker image, will try to use local image if available", "image", img, "error", err)
	} else {
		_, _ = io.Copy(io.Discard, reader)
		log.Info("Successfully pulled docker image", "image", img)
	}

	containerConfig := &container.Config{
		Image:      img,
		Cmd:        append([]string{command}, args...),
		WorkingDir: workingDir,
		Env:        env,
		Tty:        false,
	}

	hostConfig := &container.HostConfig{}
	if e.containerEnv.GetVolumes() != nil {
		for dest, src := range e.containerEnv.GetVolumes() {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: src,
				Target: dest,
			})
		}
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, e.containerEnv.GetName())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
		}
		return nil, nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer close(exitCodeChan)
		defer func() {
			if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
				log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
			}
			_ = cli.Close()
		}()
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				log.Error("Error waiting for container", "error", err)
				exitCodeChan <- -1
			}
		case status := <-statusCh:
			exitCodeChan <- int(status.StatusCode)
		}
	}()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	go func() {
		defer func() { _ = out.Close() }()
		_, err = stdcopy.StdCopy(stdoutWriter, stderrWriter, out)
		if err != nil {
			log.Error("Failed to demultiplex docker stream", "error", err)
		}
		_ = stdoutWriter.Close()
		_ = stderrWriter.Close()
	}()

	return stdoutReader, stderrReader, exitCodeChan, nil
}

func (e *dockerExecutor) ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	log := logging.GetLogger()
	cli, err := newDockerClientFunc()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	img := e.containerEnv.GetImage()
	reader, err := cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		log.Warn("Failed to pull docker image, will try to use local image if available", "image", img, "error", err)
	} else {
		_, _ = io.Copy(io.Discard, reader)
		log.Info("Successfully pulled docker image", "image", img)
	}

	containerConfig := &container.Config{
		Image:        img,
		Cmd:          append([]string{command}, args...),
		WorkingDir:   workingDir,
		Env:          env,
		Tty:          false,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{}
	if e.containerEnv.GetVolumes() != nil {
		for dest, src := range e.containerEnv.GetVolumes() {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: src,
				Target: dest,
			})
		}
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, e.containerEnv.GetName())
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create container: %w", err)
	}

	attachResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to attach to container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
		}
		return nil, nil, nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer close(exitCodeChan)
		defer func() {
			if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
				log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
			}
			_ = cli.Close()
		}()
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				log.Error("Error waiting for container", "error", err)
				exitCodeChan <- -1
			}
		case status := <-statusCh:
			exitCodeChan <- int(status.StatusCode)
		}
	}()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	go func() {
		defer func() { _ = stdoutWriter.Close() }()
		defer func() { _ = stderrWriter.Close() }()
		_, err = stdcopy.StdCopy(stdoutWriter, stderrWriter, attachResp.Reader)
		if err != nil {
			log.Error("Failed to demultiplex docker stream", "error", err)
		}
	}()

	return &closeWriter{conn: attachResp.Conn}, stdoutReader, stderrReader, exitCodeChan, nil
}
type closeWriter struct {
	conn net.Conn
}

func (c *closeWriter) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}

func (c *closeWriter) Close() error {
	if cw, ok := c.conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return c.conn.Close()
}
