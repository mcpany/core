
package util

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestIsDockerSocketAccessible(t *testing.T) {
	originalFunc := IsDockerSocketAccessibleFunc
	defer func() { IsDockerSocketAccessibleFunc = originalFunc }()

	t.Run("accessible", func(t *testing.T) {
		IsDockerSocketAccessibleFunc = func() bool {
			return true
		}
		assert.True(t, IsDockerSocketAccessible())
	})

	t.Run("inaccessible", func(t *testing.T) {
		IsDockerSocketAccessibleFunc = func() bool {
			return false
		}
		assert.False(t, IsDockerSocketAccessible())
	})
}

func TestCloseDockerClient(t *testing.T) {
	// This is a smoke test to ensure CloseDockerClient doesn't panic.
	// A proper test would require refactoring to use interfaces.
	originalClient := dockerClient
	defer func() { dockerClient = originalClient }()

	dockerClient = nil
	CloseDockerClient() // Should not panic

	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	assert.NoError(t, err)
	CloseDockerClient() // Should not panic
}

func TestIsDockerSocketAccessibleDefault(t *testing.T) {
	originalClient := dockerClient
	originalOnce := once
	originalInit := initDockerClient

	defer func() {
		dockerClient = originalClient
		once = originalOnce
		initDockerClient = originalInit
	}()

	t.Run("ping success", func(t *testing.T) {
		once = sync.Once{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", "1.41")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		initDockerClient = func() {
			var err error
			dockerClient, err = client.NewClientWithOpts(
				client.WithHost(server.URL),
				client.WithHTTPClient(server.Client()),
				client.WithAPIVersionNegotiation(),
			)
			assert.NoError(t, err)
		}

		assert.True(t, isDockerSocketAccessibleDefault())
	})

	t.Run("ping failure", func(t *testing.T) {
		once = sync.Once{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		initDockerClient = func() {
			var err error
			dockerClient, err = client.NewClientWithOpts(
				client.WithHost(server.URL),
				client.WithHTTPClient(server.Client()),
				client.WithAPIVersionNegotiation(),
			)
			assert.NoError(t, err)
		}
		assert.False(t, isDockerSocketAccessibleDefault())
	})

	t.Run("client creation failure", func(t *testing.T) {
		once = sync.Once{}
		initDockerClient = func() {
			dockerClient = nil
		}
		assert.False(t, isDockerSocketAccessibleDefault())
	})
}
