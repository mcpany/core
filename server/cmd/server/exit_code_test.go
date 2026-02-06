// Package main tests the server exit codes.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type mockFailingRunner struct{}

func (m *mockFailingRunner) Run(opts app.RunOptions) error {
	if opts.ShutdownTimeout != 10*time.Second {
		return fmt.Errorf("expected shutdown timeout of 10s, but got %v", opts.ShutdownTimeout)
	}
	return errors.New("mock run failure")
}

func (m *mockFailingRunner) RunHealthServer(_ string) error {
	return nil
}

func (m *mockFailingRunner) ReloadConfig(_ context.Context, _ afero.Fs, _ []string) error {
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
