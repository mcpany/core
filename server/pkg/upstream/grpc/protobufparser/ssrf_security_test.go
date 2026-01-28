package protobufparser_test

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	"github.com/stretchr/testify/assert"
)

func TestParseProtoByReflection_SSRF_Protection(t *testing.T) {
	// Start a dummy listener on localhost
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()
	addr := l.Addr().String()

	t.Logf("Dummy listener on %s", addr)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Attempt to connect via reflection
	_, err = protobufparser.ParseProtoByReflection(ctx, addr)

	// In a secure system, it should fail immediately with "ssrf attempt blocked".

	if err == nil {
		t.Fatal("Vulnerability: Connected successfully (SSRF protection failed)")
	} else {
		t.Logf("Got error: %v", err)
		if strings.Contains(err.Error(), "ssrf attempt blocked") {
			t.Log("Success: SSRF protection blocked the attempt.")
		} else {
			assert.Fail(t, "Failed with unrelated error (not SSRF block): "+err.Error())
		}
	}
}
