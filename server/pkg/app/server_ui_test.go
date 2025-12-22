package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *Application {
	return NewApplication()
}

func TestRun_UI(t *testing.T) {
	// Create temporary ui directory in current package dir
	err := os.Mkdir("ui", 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Failed to create ui dir: %v", err)
	}
	defer os.RemoveAll("ui")

	err = os.WriteFile("ui/index.html", []byte("<html><body>Hello</body></html>"), 0644)
	require.NoError(t, err)

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := newTestApp(t)

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	errChan := make(chan error, 1)

	go func() {
		// Run server
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", port), "localhost:0", nil, 5*time.Second)
	}()

	// Verify UI endpoint
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/ui/")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "UI endpoint should be reachable")

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
