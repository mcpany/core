/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package command provides interfaces and implementations for executing commands.
package command

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Executor is an interface for executing commands.
type Executor interface {
	Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
	ExecuteWithStdIO(ctx context.Context, command string, args []string, workingDir string, env []string) (stdin io.WriteCloser, stdout, stderr io.ReadCloser, exitCode <-chan int, err error)
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

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	exitCodeChan := make(chan int, 1)
	go func() {
		defer close(exitCodeChan)
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

	return stdin, stdout, stderr, exitCodeChan, nil
}

type dockerExecutor struct {
	containerEnv *configv1.ContainerEnvironment
}

func newDockerExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	return &dockerExecutor{containerEnv: containerEnv}
}

func (e *dockerExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	log := logging.GetLogger()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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

	return attachResp.Conn, stdoutReader, stderrReader, exitCodeChan, nil
}
