// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type mockFailingRunner struct{}

func (m *mockFailingRunner) Run(
	ctx context.Context,
	fs afero.Fs,
	stdio bool,
	jsonrpcPort string,
	registrationPort string,
	configPaths []string,
	shutdownTimeout time.Duration,
) error {
	if shutdownTimeout != 10*time.Second {
		return fmt.Errorf("expected shutdown timeout of 10s, but got %v", shutdownTimeout)
	}
	return errors.New("mock run failure")
}

func (m *mockFailingRunner) RunHealthServer(jsonrpcPort string) error {
	return nil
}

func (m *mockFailingRunner) ReloadConfig(fs afero.Fs, configPaths []string) error {
	return nil
}

var _ app.Runner = &mockFailingRunner{}

func TestMain_FailingExitCode(t *testing.T) {
	if os.Getenv("GO_TEST_EXIT_CODE") == "1" {
		appRunner = &mockFailingRunner{}
		os.Args = append(os.Args, "--shutdown-timeout=10s")
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMain_FailingExitCode$") //nolint:gosec
	cmd.Env = append(os.Environ(), "GO_TEST_EXIT_CODE=1")

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	assert.True(t, ok, "err should be of type *exec.ExitError")
	assert.False(t, e.Success(), "process should exit with a non-zero status code")
}
