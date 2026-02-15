package wasm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockWASMRuntime(t *testing.T) {
	runtime := NewMockRuntime()
	defer func() {
		assert.NoError(t, runtime.Close())
	}()

	// Test Load
	plugin, err := runtime.LoadPlugin(context.Background(), []byte("fake-wasm"))
	assert.NoError(t, err)
	assert.NotNil(t, plugin)
	defer func() {
		assert.NoError(t, plugin.Close())
	}()

	// Test Execute
	out, err := plugin.Execute(context.Background(), "run")
	assert.NoError(t, err)
	assert.Equal(t, []byte("success"), out)

	// Test Error
	_, err = plugin.Execute(context.Background(), "error")
	assert.Error(t, err)

	// Test Empty Bytecode
	_, err = runtime.LoadPlugin(context.Background(), nil)
	assert.Error(t, err)
}
