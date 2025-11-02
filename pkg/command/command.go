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

package command

import (
	"bytes"
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
	Execute(ctx context.Context, command string, args []string, workingDir string, env []string) (stdout, stderr []byte, exitCode int, err error)
}

// NewExecutor creates a new command executor.
func NewExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	if containerEnv != nil && containerEnv.GetImage() != "" {
		return newDockerExecutor(containerEnv)
	}
	return &localExecutor{}
}

type localExecutor struct{}

func (e *localExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) ([]byte, []byte, int, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDir
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Indicates an error other than a non-zero exit code
		}
	}

	return stdout.Bytes(), stderr.Bytes(), exitCode, err
}

type dockerExecutor struct {
	containerEnv *configv1.ContainerEnvironment
}

func newDockerExecutor(containerEnv *configv1.ContainerEnvironment) Executor {
	return &dockerExecutor{containerEnv: containerEnv}
}

func (e *dockerExecutor) Execute(ctx context.Context, command string, args []string, workingDir string, env []string) ([]byte, []byte, int, error) {
	log := logging.GetLogger()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, -1, fmt.Errorf("failed to create docker client: %w", err)
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
	if e.containerEnv.GetMounts() != nil {
		for src, dest := range e.containerEnv.GetMounts() {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: src,
				Target: dest,
			})
		}
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, e.containerEnv.GetName())
	if err != nil {
		return nil, nil, -1, fmt.Errorf("failed to create container: %w", err)
	}
	defer func() {
		if rmErr := cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			log.Error("Failed to remove container", "containerID", resp.ID, "error", rmErr)
		}
	}()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, nil, -1, fmt.Errorf("failed to start container: %w", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, nil, -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			return nil, nil, int(status.StatusCode), fmt.Errorf("failed to get container logs: %w", err)
		}
		defer out.Close()

		var stdout, stderr bytes.Buffer
		_, err = stdcopy.StdCopy(&stdout, &stderr, out)
		if err != nil {
			return nil, nil, int(status.StatusCode), fmt.Errorf("failed to demultiplex docker stream: %w", err)
		}
		return stdout.Bytes(), stderr.Bytes(), int(status.StatusCode), nil
	case <-ctx.Done():
		return nil, nil, -1, ctx.Err()
	}
	return nil, nil, -1, fmt.Errorf("unexpected error in docker execution")
}
