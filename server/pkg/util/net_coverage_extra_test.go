// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListenWithRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Success case
	lis, err := ListenWithRetry(ctx, "tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	assert.NotNil(t, lis)
	defer lis.Close()

	// 2. Failure case (invalid address)
	lis2, err2 := ListenWithRetry(ctx, "tcp", "invalid-address")
	assert.Error(t, err2)
	assert.Nil(t, lis2)

	// 3. Busy port (triggers retry logic)
	// We bind a port, then try to bind it again.
	lis3, err3 := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err3)
	defer lis3.Close()
	addr := lis3.Addr().String()

	// Try to listen on the same address
	// This will fail with EADDRINUSE and retry.
	// Since we don't release the port, it will eventually fail after maxRetries.
	// address doesn't end in :0, so maxRetries is 1.
	// Wait, we need maxRetries > 1 to test backoff.
	// But maxRetries is 1 unless address ends in :0.
	// If I pass "127.0.0.1:0", I get a NEW port. I can't force collision on "127.0.0.1:0".
	// Collision happens on specific port e.g. "127.0.0.1:12345".
	// But then suffix is not ":0".
	// So maxRetries is 1.

	// Code:
	// maxRetries := 1
	// if strings.HasSuffix(address, ":0") { maxRetries = 10 }

	// So I can't test retry loop with specific port unless I change the code or trick it.
	// I cannot trick HasSuffix.

	// So ListenWithRetry ONLY retries for dynamic port allocation (:0).
	// But dynamic port allocation rarely fails with EADDRINUSE unless system is exhausted.
	// So the retry logic is specifically for "OS gave me a port but said it's in use" race condition?
	// Or maybe "I asked for port 0, OS gave me X, but X is in use"?
	// No, `Listen(..., ":0")` returns a listener on a random free port. It shouldn't fail with EADDRINUSE.
	// Maybe on some systems it does?

	// If I cannot trigger the retry loop, I cannot cover it.
	// But the code explicitly checks suffix ":0".

	lis4, err4 := ListenWithRetry(ctx, "tcp", addr)
	assert.Error(t, err4)
	assert.Nil(t, lis4)
}
