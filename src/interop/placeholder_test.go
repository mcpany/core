package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

// TestPlaceholderAdapter verifies that the PlaceholderAdapter behaves exactly as an unimplemented stub.
//
// Summary: Validates that a constructed PlaceholderAdapter correctly identifies its name and returns Not Implemented errors for task execution.
//
// Parameters:
//   - t (*testing.T): The testing framework instance used for assertions and failure reporting.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

    stream, err := adapter.StreamTask(context.Background(), &interop.Task{})
    if err == nil {
        t.Error("Expected StreamTask to return an error, got nil")
    } else if err.Error() != "placeholder method: not implemented" {
        t.Errorf("Expected correct StreamTask error, got: %s", err.Error())
    }
    if stream != nil {
        t.Error("Expected StreamTask stream to be nil")
    }
}

// TestNewPlaceholderAdapterNilCapabilities verifies the initialization behavior when providing nil capabilities.
//
// Summary: Ensures that NewPlaceholderAdapter safely initializes an empty map instead of panicking when nil is passed.
//
// Parameters:
//   - t (*testing.T): The testing framework instance used for assertions and failure reporting.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func TestNewPlaceholderAdapterNilCapabilities(t *testing.T) {
    adapter := interop.NewPlaceholderAdapter("Nil Caps Hub", nil)

    if adapter.Name() != "Nil Caps Hub" {
        t.Errorf("Expected name 'Nil Caps Hub', got '%s'", adapter.Name())
    }

    if adapter.SupportsCapability("any_capability") {
        t.Error("Expected no capabilities to be supported when initialized with nil")
    }
}

// TestRegisterPlaceholders verifies the mass registration of all known unimplemented roadmap features.
//
// Summary: Tests that RegisterPlaceholders successfully adds all placeholder services (e.g. CI/CD Cache Integrity Guard) to the AdapterHub.
//
// Parameters:
//   - t (*testing.T): The testing framework instance used for assertions and failure reporting.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the provided test AdapterHub instance by adding multiple placeholder adapters.
func TestRegisterPlaceholders(t *testing.T) {
    hub := interop.NewAdapterHub()
    interop.RegisterPlaceholders(hub)

    // Test a sample feature that should be registered
    task := &interop.Task{
        Framework: "Dynamic Mesh Resilience (DMR) Hub",
        Intent: "any",
    }

    _, err := hub.RouteTask(context.Background(), task)
    if err == nil {
        t.Error("Expected routing to placeholder to return an error")
    } else if err.Error() != "Not Implemented: Dynamic Mesh Resilience (DMR) Hub is a placeholder service" {
        t.Errorf("Unexpected error from routed placeholder: %s", err.Error())
    }

    // Test that something NOT in the missingFeatures list returns an adapter not registered error
    taskUnknown := &interop.Task{
        Framework: "Not A Real Feature (NARF)",
        Intent: "any",
    }
    _, err = hub.RouteTask(context.Background(), taskUnknown)
    if err == nil {
        t.Error("Expected unknown framework to return an error")
    } else if err.Error() != "no adapter registered for framework: Not A Real Feature (NARF)" {
        t.Errorf("Unexpected error for unknown framework: %s", err.Error())
    }
}
