// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateIP_Coverage(t *testing.T) {
	// 169.254.0.1 is link-local. IsPrivateIP checks linkLocalBlocks.
	if !IsPrivateIP(net.ParseIP("169.254.0.1")) {
		t.Error("169.254.0.1 should be private IP")
	}

	// 127.0.0.1 is loopback.
	if !IsPrivateIP(net.ParseIP("127.0.0.1")) {
		t.Error("127.0.0.1 should be private IP")
	}

	// 10.0.0.1 is private network.
	if !IsPrivateIP(net.ParseIP("10.0.0.1")) {
		t.Error("10.0.0.1 should be private IP")
	}

	// 8.8.8.8 is public.
	if IsPrivateIP(net.ParseIP("8.8.8.8")) {
		t.Error("8.8.8.8 should not be private IP")
	}
}
