package interop

import (
	"context"
	"testing"
)

func TestHACAProvider(t *testing.T) {
	baseAdapter := NewOpenClawAdapter()
	hacaProvider := NewHACAProvider(baseAdapter, "TPM-001")

	if hacaProvider.Name() != "OpenClaw-HACA" {
		t.Errorf("expected OpenClaw-HACA, got %s", hacaProvider.Name())
	}

	if !hacaProvider.SupportsCapability("adaptive_reasoning") {
		t.Errorf("expected HACA provider to support adaptive_reasoning")
	}

	task := &Task{
		ID:        "task-1",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload: map[string]string{
			"lineage_id": "parent-task-123",
		},
	}

	result, err := hacaProvider.HandleTask(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Telemetry["haca_lineage_id"] != "parent-task-123" {
		t.Errorf("expected lineage_id parent-task-123, got %s", result.Telemetry["haca_lineage_id"])
	}

	if result.Telemetry["haca_enclave_id"] != "TPM-001" {
		t.Errorf("expected enclave_id TPM-001, got %s", result.Telemetry["haca_enclave_id"])
	}

	if result.Telemetry["haca_signature"] == "" {
		t.Errorf("expected non-empty haca_signature")
	}

	if result.Telemetry["haca_compute_millis"] == "" || result.Telemetry["haca_compute_millis"] == "0" {
		t.Errorf("expected positive haca_compute_millis")
	}

	if result.Telemetry["haca_tokens_used"] == "" || result.Telemetry["haca_tokens_used"] == "0" {
		t.Errorf("expected positive haca_tokens_used")
	}
}
