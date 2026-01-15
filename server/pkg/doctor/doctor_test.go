package doctor_test

import (
	"context"
	"fmt"
	"testing"
	"time"
    "net"

	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestPortCheck_Available(t *testing.T) {
	// Use a random port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // Close it so it's available

	check := doctor.NewPortCheck(fmt.Sprintf(":%d", port))
	// Wait a bit to ensure OS releases it
	time.Sleep(10 * time.Millisecond)

	res := check.Run(context.Background())
	assert.Equal(t, doctor.Info, res.Severity, "Message: %s", res.Message)
}

func TestPortCheck_Busy(t *testing.T) {
	// Occupy a port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	check := doctor.NewPortCheck(fmt.Sprintf(":%d", port))
	res := check.Run(context.Background())

	assert.Equal(t, doctor.Error, res.Severity)
	assert.Contains(t, res.Message, "not available")
}

func TestConfigCheck_Missing(t *testing.T) {
	fs := afero.NewMemMapFs()
	check := doctor.NewConfigCheck(fs, []string{"missing.yaml"})

	res := check.Run(context.Background())
	// Should fail because file is missing (LoadResolvedConfig fails)
	assert.Equal(t, doctor.Error, res.Severity)
}

func TestConfigCheck_Valid(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Fix: config schema expects "upstream_services" not "services"
	// and usually wrapped in a root object if checking specific file format.
	// But LoadResolvedConfig assumes root object.
	// The proto definition likely has upstream_services field.
	// Let's use empty object or "upstream_services: []"
	err := afero.WriteFile(fs, "config.yaml", []byte("upstream_services: []"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	check := doctor.NewConfigCheck(fs, []string{"config.yaml"})
	res := check.Run(context.Background())

	assert.Equal(t, doctor.Info, res.Severity, "Message: %s", res.Message)
}
