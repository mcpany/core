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
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"path/filepath"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/validation"
)

// Executor defines the interface for running system commands, either locally or within a container.
type Executor interface {
	// Execute runs a command and returns its standard output and standard error as streams.
	// It is useful for non-interactive commands where input is not required.
	//
	// Parameters:
	//   ctx: The context for managing the command's lifecycle (e.g., cancellation).
	//   command: The executable to run.
	//   args: A slice of arguments to pass to the executable.
	//   workingDir: The directory in which to execute the command.
	//   env: A slice of environment variables in "KEY=VALUE" format.
	//
	// Returns:
	//   stdout: A stream for reading the command's standard output.
	//   stderr: A stream for reading the command's standard error.
	//   exitCode: A channel that receives the exit code when the command finishes.
	//   err: An error if the command fails to start.
	Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (stdout, stderr io.ReadCloser, exitCode <-chan int, err error)

	// ExecuteWithStdIO runs a command and provides access to stdin, stdout, and stderr.
	// It supports interactive commands where writing to stdin is necessary.
	//
	// Parameters:
	//   ctx: The context for managing the command's lifecycle.
	//   command: The executable to run.
	//   args: A slice of arguments to pass to the executable.
	//   workingDir: The directory in which to execute the command.
	//   env: A slice of environment variables in "KEY=VALUE" format.
	//
	// Returns:
	//   stdin: A stream for writing to the command's standard input.
	//   stdout: A stream for reading the command's standard output.
	//   stderr: A stream for reading the command's standard error.
	//   exitCode: A channel that receives the exit code when the command finishes.
	//   err: An error if the command fails to start.
	ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (stdin io.WriteCloser, stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
}

// NewExecutor creates a new command executor based on the provided environment configuration.
// If a container environment is specified with an image, it returns a Docker executor.
// Otherwise, it returns a local executor.
//
// Parameters:
//   containerEnv: The configuration for the container environment (optional).
//
// Returns:
//   An implementation of the Executor interface.
func NewExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	if containerEnv != nil && containerEnv.GetImage() != "" {
		return newDockerExecutor(containerEnv)
	}
	return &localExecutor{}
}

// NewLocalExecutor creates a new executor that runs commands directly on the host system.
//
// Returns:
//   An Executor implementation for local execution.
func NewLocalExecutor() Executor {
	return &localExecutor{}
}

type localExecutor struct{}

// Execute executes a command locally on the host system.
//
// Parameters:
//   ctx: The context for the request.
//   command: The command to execute.
//   args: Arguments for the command.
//   workingDir: The directory to execute in.
//   env: Environment variables.
//
// Returns:
//   stdout: Standard output stream.
//   stderr: Standard error stream.
//   exitCode: Channel for exit code.
//   error: Error if startup fails.
func (e *localExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	if workingDir != "" {
		if err := validation.IsAllowedPath(workingDir); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid working directory %q: %w", workingDir, err)
		}
	}

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

// ExecuteWithStdIO executes a command locally with full stdin/stdout/stderr support.
//
// Parameters:
//   ctx: The context for the request.
//   command: The command to execute.
//   args: Arguments for the command.
//   workingDir: The directory to execute in.
//   env: Environment variables.
//
// Returns:
//   stdin: Standard input stream.
//   stdout: Standard output stream.
//   stderr: Standard error stream.
//   exitCode: Channel for exit code.
//   error: Error if startup fails.
func (e *localExecutor) ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	if workingDir != "" {
		if err := validation.IsAllowedPath(workingDir); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("invalid working directory %q: %w", workingDir, err)
		}
	}

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDir
	cmd.Env = env

	// Use io.Pipe to avoid race condition where cmd.Wait() closes pipes before we are done reading
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
		defer func() { _ = stdoutW.Close() }()
		defer func() { _ = stderrW.Close() }()
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
	containerEnv  *configv1.ContainerEnvironment
	clientFactory func() (DockerClient, error)
}

func newDockerExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	return &dockerExecutor{
		containerEnv: containerEnv,
		clientFactory: func() (DockerClient, error) {
			return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		},
	}
}

// Execute runs a command inside a Docker container.
// It handles pulling the image, creating the container, and streaming logs.
//
// Parameters:
//   ctx: The context for the request.
//   command: The command to execute.
//   args: Arguments for the command.
//   workingDir: The directory inside the container to execute in.
//   env: Environment variables.
//
// Returns:
//   stdout: Standard output stream.
//   stderr: Standard error stream.
//   exitCode: Channel for exit code.
//   error: Error if container setup or execution fails.
func (e *dockerExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	log := logging.GetLogger()
	cli, err := e.clientFactory()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	// We pass ownership of the client to the goroutine waiting for the container.
	// defer func() { _ = cli.Close() }()

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
			// Validate host path (dest) to prevent mounting sensitive directories
			if err := validation.IsAllowedPath(dest); err != nil {
				_ = cli.Close()
				return nil, nil, nil, fmt.Errorf("invalid volume mount source %q: %w", dest, err)
			}

			// Docker requires absolute path for bind mounts
			absDest, err := filepath.Abs(dest)
			if err != nil {
				_ = cli.Close()
				return nil, nil, nil, fmt.Errorf("failed to resolve absolute path for %q: %w", dest, err)
			}

			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: absDest,
				Target: src,
			})
		}
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, e.containerEnv.GetName())
	if err != nil {
		_ = cli.Close()
		return nil, nil, nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
		}
		_ = cli.Close()
		return nil, nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		_ = cli.Close()
		return nil, nil, nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to wait for container exit and close client when everything is done
	go func() {
		wg.Wait()
		_ = cli.Close()
	}()

	exitCodeChan := make(chan int, 1)
	go func() {
		defer wg.Done()
		defer close(exitCodeChan)
		defer func() {
			if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
				log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
			}
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
		defer wg.Done()
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

// ExecuteWithStdIO runs a command inside a Docker container with interactive I/O.
//
// Parameters:
//   ctx: The context for the request.
//   command: The command to execute.
//   args: Arguments for the command.
//   workingDir: The directory inside the container to execute in.
//   env: Environment variables.
//
// Returns:
//   stdin: Standard input stream.
//   stdout: Standard output stream.
//   stderr: Standard error stream.
//   exitCode: Channel for exit code.
//   error: Error if container setup or execution fails.
func (e *dockerExecutor) ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	log := logging.GetLogger()
	cli, err := e.clientFactory()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	// We pass ownership of the client to the goroutine waiting for the container.
	// defer func() { _ = cli.Close() }()

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
			// Validate host path (dest) to prevent mounting sensitive directories
			if err := validation.IsAllowedPath(dest); err != nil {
				_ = cli.Close()
				return nil, nil, nil, nil, fmt.Errorf("invalid volume mount source %q: %w", dest, err)
			}

			// Docker requires absolute path for bind mounts
			absDest, err := filepath.Abs(dest)
			if err != nil {
				_ = cli.Close()
				return nil, nil, nil, nil, fmt.Errorf("failed to resolve absolute path for %q: %w", dest, err)
			}

			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: absDest,
				Target: src,
			})
		}
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, e.containerEnv.GetName())
	if err != nil {
		_ = cli.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to create container: %w", err)
	}

	attachResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		_ = cli.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to attach to container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
		}
		_ = cli.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer func() { _ = cli.Close() }() // Close client when monitoring is done
		defer close(exitCodeChan)
		defer func() {
			if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
				log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
			}
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

// Write writes the provided byte slice to the underlying connection.
//
// Parameters:
//   p: The data to write.
//
// Returns:
//   n: The number of bytes written.
//   err: An error if the write fails.
func (c *closeWriter) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}

// Close closes the write side of the connection, signaling EOF to the peer.
//
// Returns:
//   An error if closing fails.
func (c *closeWriter) Close() error {
	if cw, ok := c.conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return c.conn.Close()
}
