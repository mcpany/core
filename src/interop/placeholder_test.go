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

func TestNewPlaceholderAdapterNilCapabilities(t *testing.T) {
    adapter := interop.NewPlaceholderAdapter("Nil Caps Hub", nil)

    if adapter.Name() != "Nil Caps Hub" {
        t.Errorf("Expected name 'Nil Caps Hub', got '%s'", adapter.Name())
    }

    if adapter.SupportsCapability("any_capability") {
        t.Error("Expected no capabilities to be supported when initialized with nil")
    }
}

func TestVerifyStylometricSignature(t *testing.T) {
    adapter := interop.NewPlaceholderAdapter("GIMM Hub", nil)

    err := adapter.VerifyStylometricSignature(context.Background(), "")
    if err == nil {
        t.Error("Expected error for empty trace, got nil")
    } else if err.Error() != "invalid stylometric signature: trace is empty" {
        t.Errorf("Expected 'invalid stylometric signature: trace is empty', got '%s'", err.Error())
    }

    longTrace := string(make([]byte, 1001))
    err = adapter.VerifyStylometricSignature(context.Background(), longTrace)
    if err == nil {
        t.Error("Expected error for long trace, got nil")
    } else if err.Error() != "invalid stylometric signature: trace exceeds maximum length" {
        t.Errorf("Expected 'invalid stylometric signature: trace exceeds maximum length', got '%s'", err.Error())
    }

    err = adapter.VerifyStylometricSignature(context.Background(), "valid trace")
    if err != nil {
        t.Errorf("Expected nil for valid trace, got: %v", err)
    }
}

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
