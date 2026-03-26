// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	buspb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MainTestSuite ...
// Summary: MainTestSuite
	suite.Suite
}

// TestSetup_InMemoryBus ...
// Summary: TestSetup_InMemoryBus
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Ensure REDIS_ADDR is not set
	_ = os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	s.NoError(err, "setup() should not return an error when using in-memory bus")
}

// TestSetup_ValidRedisAddress ...
// Summary: TestSetup_ValidRedisAddress
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	s.T().Setenv("REDIS_ADDR", "127.0.0.1:6379")

	_, err := setup()
	s.NoError(err, "setup() should not return an error with a valid REDIS_ADDR because the connection is lazy")
}

// TestSetup_InvalidRedisAddress ...
// Summary: TestSetup_InvalidRedisAddress
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	s.T().Setenv("REDIS_ADDR", "invalid-address")

	_, err := setup()
	s.NoError(err, "setup() should not return an error with an invalid REDIS_ADDR because the connection is lazy")
}

// TestMainLifecycle ...
// Summary: TestMainLifecycle
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Ensure REDIS_ADDR is not set, so we use the in-memory bus
	_ = os.Unsetenv("REDIS_ADDR")

	mainDone := make(chan struct{})
	go func() {
		defer close(mainDone)
		main()
	}()

	// Allow some time for the worker to start
	time.Sleep(100 * time.Millisecond)

	// Send an interrupt signal to trigger shutdown
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	// Wait for main to exit
	select {
	case <-mainDone:
		// main exited gracefully
	case <-time.After(2 * time.Second):
		s.T().Fatal("main function did not exit in time")
	}
}

// TestMainTestSuite ...
// Summary: TestMainTestSuite
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	suite.Run(t, new(MainTestSuite))
}

// TestSetup_InMemoryBus ...
// Summary: TestSetup_InMemoryBus
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Unset REDIS_ADDR to fall back to in-memory bus
	_ = os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	assert.NoError(t, err, "setup() should not return an error when using in-memory bus")
}

// TestSetup_DirectValidationError ...
// Summary: TestSetup_DirectValidationError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// This test directly manipulates the config to trigger an error in the validation logic
	// to ensure the error handling in setup() is covered.
	busConfig := &buspb.MessageBus{}
	busConfig.SetRedis(&buspb.RedisBus{}) // Set an empty redis bus to trigger a validation error

	_, err := setupWithConfig(busConfig)
	assert.Error(t, err, "setupWithConfig() should return an error with an invalid busConfig")
}
