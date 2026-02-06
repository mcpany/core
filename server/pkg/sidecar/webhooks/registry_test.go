package webhooks

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockHandler struct{}

func (h *mockHandler) Handle(w http.ResponseWriter, r *http.Request) {}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	t.Run("Register and Get", func(t *testing.T) {
		h := &mockHandler{}
		r.Register("test", h)

		got, ok := r.Get("test")
		assert.True(t, ok)
		assert.Equal(t, h, got)

		_, ok = r.Get("missing")
		assert.False(t, ok)
	})

	t.Run("Concurrency", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				r.Register("key", &mockHandler{})
				r.Get("key")
			}(i)
		}
		wg.Wait()
	})
}
