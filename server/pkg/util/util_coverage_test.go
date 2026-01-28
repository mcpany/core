// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedactMap_Coverage(t *testing.T) {
	// Nested map copying coverage
	nested := map[string]interface{}{
		"token": "secret",
	}
	original := map[string]interface{}{
		"safe": "value",
		"nested": nested,
	}

	redacted := RedactMap(original)

	// Verify redaction happened
	assert.NotEqual(t, original, redacted)

	// Verify nested is redacted
	nestedRes := redacted["nested"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", nestedRes["token"])

	// Verify deep copy (modifying redacted shouldn't affect original)
	nestedRes["safe_nested"] = "added"
	originalNested := original["nested"].(map[string]interface{})
	_, ok := originalNested["safe_nested"]
	assert.False(t, ok, "Original map should not be modified")

	// Nested slice copying coverage
	nestedSlice := []interface{}{
		map[string]interface{}{"token": "secret"},
	}
	originalSliceMap := map[string]interface{}{
		"list": nestedSlice,
	}

	redactedSliceMap := RedactMap(originalSliceMap)
	listRes := redactedSliceMap["list"].([]interface{})
	item := listRes[0].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", item["token"])
}

func TestListenWithRetry_Coverage(t *testing.T) {
	// 1. Failure case (non-port 0)
	// Bind a port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Could not bind port")
	}
	defer l.Close()
	addr := l.Addr().String()

	// Try to bind same port using ListenWithRetry
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = ListenWithRetry(ctx, "tcp", addr)
	assert.Error(t, err)
	// Should contain "address already in use" or similar
	assert.True(t, strings.Contains(err.Error(), "bind") || strings.Contains(err.Error(), "in use") || strings.Contains(err.Error(), "only one usage"))
}
