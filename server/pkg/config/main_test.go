package config

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Create dummy cert file for tests
	_ = os.WriteFile("dummy-cert.pem", []byte("cert"), 0600)
	code := m.Run()
	_ = os.Remove("dummy-cert.pem")
	os.Exit(code)
}
