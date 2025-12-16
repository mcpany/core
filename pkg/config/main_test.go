package config

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Create dummy cert file for tests
	_ = os.WriteFile("dummy-cert.pem", []byte("cert"), 0644)
	code := m.Run()
	os.Remove("dummy-cert.pem")
	os.Exit(code)
}
