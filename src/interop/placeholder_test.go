package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestPlaceholderAdapter(t *testing.T) {
	adapter := interop.NewPlaceholderAdapter("Dynamic Mesh Resilience (DMR) Hub", map[string]bool{"shard_migration": true})

	if adapter.Name() != "Dynamic Mesh Resilience (DMR) Hub" {
		t.Errorf("Expected name 'Dynamic Mesh Resilience (DMR) Hub', got '%s'", adapter.Name())
	}

	if !adapter.SupportsCapability("shard_migration") {
		t.Error("Expected capability 'shard_migration' to be supported")
	}

	if adapter.SupportsCapability("unsupported") {
		t.Error("Expected capability 'unsupported' to be not supported")
	}

	_, err := adapter.HandleTask(context.Background(), &interop.Task{})
	if err == nil {
		t.Error("Expected HandleTask to return an error, got nil")
	} else if err.Error() != "Not Implemented: Dynamic Mesh Resilience (DMR) Hub is a placeholder service" {
		t.Errorf("Expected correct HandleTask error, got: %s", err.Error())
	}

	err = adapter.SyncMemoryShard(context.Background(), &interop.MemoryShard{})
	if err == nil {
		t.Error("Expected SyncMemoryShard to return an error, got nil")
	} else if err.Error() != "Not Implemented: Dynamic Mesh Resilience (DMR) Hub is a placeholder service" {
		t.Errorf("Expected correct SyncMemoryShard error, got: %s", err.Error())
	}
}
