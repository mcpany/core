package util

import (
	"context"
	"net"
	"testing"
)

func TestCheckConnection_SSRF_Protection(t *testing.T) {
	// Start a local listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()

	// 1. Verify CheckConnection FAILS by default (Correct behavior: Blocks SSRF)
	err = CheckConnection(context.Background(), addr)
	if err == nil {
		t.Fatalf("CheckConnection succeeded unexpectedly! It should have blocked the loopback connection.")
	} else {
		t.Logf("CheckConnection correctly blocked connection by default: %v", err)
	}

	// 2. Verify SafeDialContext FAILS by default (Correct behavior)
	_, err = SafeDialContext(context.Background(), "tcp", addr)
	if err == nil {
		t.Errorf("SafeDialContext succeeded unexpectedly (it should block loopback by default)")
	}

	// 3. Verify CheckConnection SUCCEEDS when explicitly allowed via Env Var
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	err = CheckConnection(context.Background(), addr)
	if err != nil {
		t.Errorf("CheckConnection failed unexpectedly when MCPANY_ALLOW_LOOPBACK_RESOURCES=true: %v", err)
	} else {
		t.Logf("CheckConnection succeeded when allowed via env var.")
	}
}
