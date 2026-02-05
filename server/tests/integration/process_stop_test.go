// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestManagedProcess_StopFallback tests that the ManagedProcess.Stop function
// correctly falls back to sending SIGINT to the process directly if sending to the
// process group fails.
func TestManagedProcess_StopFallback(t *testing.T) {
	// This test uses a simple 'sleep' command. We will patch the syscall.Kill
	// function to simulate a failure when trying to signal the process group.
	// This will allow us to test the fallback logic in a controlled manner.
	cmdPath, err := exec.LookPath("sleep")
	require.NoError(t, err, "sleep command not found in PATH")

	mp := NewManagedProcess(t, "StopFallbackTest", cmdPath, []string{"30"}, nil)
	err = mp.Start()
	require.NoError(t, err, "Failed to start managed process")

	// Patch syscall.Kill to simulate failure
	originalSyscallKill := syscallKill
	syscallKill = func(pid int, sig syscall.Signal) (err error) {
		// We only want to simulate the failure of the group kill
		if pid < 0 {
			return syscall.EPERM
		}
		return originalSyscallKill(pid, sig)
	}
	defer func() { syscallKill = originalSyscallKill }()

	// The Stop() function should be able to gracefully terminate the process
	// even if the group-level SIGINT fails. The test will time out if the
	// fallback logic is not implemented correctly.
	stopTimeout := 20 * time.Second
	stoppedCh := make(chan struct{})
	go func() {
		mp.Stop()
		close(stoppedCh)
	}()

	select {
	case <-stoppedCh:
		// Success: The process was stopped gracefully.
	case <-time.After(stopTimeout):
		t.Fatal("Test timed out: ManagedProcess.Stop() did not complete in time, indicating it did not fall back to single-process SIGINT.")
	}
}
