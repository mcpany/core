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
// Executor is an interface for executing commands.
//
// Summary: Represents a Executor.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Executor interface {
	// Execute executes a command and returns the stdout and stderr as streams.
	//
	// Summary: Executes a command.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - command (string): The command to execute.
	//   - args ([]string): The arguments for the command.
	//   - workingDir (string): The working directory for execution.
	//   - env ([]string): The environment variables.
	//
	// Returns:
	//   - stdout (io.ReadCloser): The standard output stream.
	//   - stderr (io.ReadCloser): The standard error stream.
	//   - exitCode (<-chan int): A channel that receives the exit code.
	//   - err (error): An error if the operation fails.
	Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
	// ExecuteWithStdIO executes a command and returns the stdin, stdout, and stderr as streams.
	//
	// Summary: Executes a command with full I/O streams.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - command (string): The command to execute.
// NewExecutor creates a new command executor.
//
// Summary: Creates a new command executor (local or docker).
// Execute executes a command locally.
//
// Summary: Executes a command on the local system.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - command (string): The command to execute.
//   - args ([]string): The arguments for the command.
//   - workingDir (string): The working directory for execution.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - env ([]string): The environment variables.
//
// Returns:
//   - io.ReadCloser: The standard output stream.
//   - io.ReadCloser: The standard error stream.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - <-chan int: A channel that receives the exit code.
//   - error: An error if the operation fails.
//
// Side Effects:
//   - Spawns a subprocess.
// Errors:
//   - triggers relevant error states on failure.
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
// ExecuteWithStdIO executes a command locally with stdin/stdout/stderr pipes.
//
// Summary: Executes a command on the local system with full I/O streams.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - command (string): The command to execute.
//   - args ([]string): The arguments for the command.
//   - workingDir (string): The working directory for execution.
//   - env ([]string): The environment variables.
//
// Returns:
//   - io.WriteCloser: The standard input stream.
//   - io.ReadCloser: The standard output stream.
//   - io.ReadCloser: The standard error stream.
//   - <-chan int: A channel that receives the exit code.
//   - error: An error if the operation fails.
//
// Side Effects:
//   - Spawns a subprocess.
// Errors:
//   - triggers relevant error states on failure.
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
// Execute executes a command inside a docker container.
//
// Summary: Executes a command inside a Docker container.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - command (string): The command to execute.
//   - args ([]string): The arguments for the command.
//   - workingDir (string): The working directory for execution.
//   - env ([]string): The environment variables.
//
// Returns:
//   - io.ReadCloser: The standard output stream.
//   - io.ReadCloser: The standard error stream.
//   - <-chan int: A channel that receives the exit code.
//   - error: An error if the operation fails.
//
// Side Effects:
//   - Creates and starts a Docker container.
// Errors:
//   - triggers relevant error states on failure.
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
// ExecuteWithStdIO executes a command inside a docker container with stdin/stdout/stderr pipes.
//
// Summary: Executes a command inside a Docker container with full I/O streams.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - command (string): The command to execute.
//   - args ([]string): The arguments for the command.
//   - workingDir (string): The working directory for execution.
//   - env ([]string): The environment variables.
//
// Returns:
//   - io.WriteCloser: The standard input stream.
//   - io.ReadCloser: The standard output stream.
//   - io.ReadCloser: The standard error stream.
//   - <-chan int: A channel that receives the exit code.
//   - error: An error if the operation fails.
//
// Side Effects:
//   - Creates and starts a Docker container.
// Errors:
//   - triggers relevant error states on failure.
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
// Write writes data to the connection.
//
// Summary: Writes data to the underlying connection.
//
// Close closes the write side of the connection.
//
// Summary: Closes the write side of the connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if closing fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Side Effects:
//   - Closes the connection writer.
// Errors:
//   - triggers relevant error states on failure.
func (c *closeWriter) Close() error {
	if cw, ok := c.conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return c.conn.Close()
}
