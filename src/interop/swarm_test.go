package interop

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSwarm(t *testing.T) {
	// Simulate an OpenClaw agent starting the task
	initialState := map[string]interface{}{
		"oc_session":   "session-001",
		"mission_root": "build_interop_bridge",
		"shards":       "initial research data",
	}

	ocData, err := json.Marshal(initialState)
	if err != nil {
		t.Fatalf("Failed to marshal OpenClaw state: %v", err)
	}

	ocBridge, err := NewBridge("OpenClaw")
	if err != nil {
		t.Fatalf("Failed to create OpenClaw bridge: %v", err)
	}

	// 1. Read context from OpenClaw
	uCtxData, err := ocBridge.ReadContext(ocData)
	if err != nil {
		t.Fatalf("Failed to read OpenClaw context: %v", err)
	}

	var uctx UniversalContext
	if err := json.Unmarshal(uCtxData, &uctx); err != nil {
		t.Fatalf("Failed to unmarshal UniversalContext: %v", err)
	}

	if uctx.SessionID != "session-001" || uctx.Intent != "build_interop_bridge" || uctx.Memory != "initial research data" {
		t.Errorf("Universal context mapping failed for OpenClaw. Got: %+v", uctx)
	}

	// 2. Hand off to CrewAI agent
	crewBridge, err := NewBridge("CrewAI")
	if err != nil {
		t.Fatalf("Failed to create CrewAI bridge: %v", err)
	}

	crewData, err := crewBridge.WriteContext(uCtxData)
	if err != nil {
		t.Fatalf("Failed to write to CrewAI context: %v", err)
	}

	var crewState map[string]interface{}
	if err := json.Unmarshal(crewData, &crewState); err != nil {
		t.Fatalf("Failed to unmarshal CrewAI state: %v", err)
	}

	expectedCrewState := map[string]interface{}{
		"crew_id":   "session-001",
		"task_goal": "build_interop_bridge",
		"context":   "initial research data",
	}

	if !reflect.DeepEqual(crewState, expectedCrewState) {
		t.Errorf("CrewAI state mapping failed. Expected %v, got %v", expectedCrewState, crewState)
	}

	// 3. CrewAI agent adds some work and we sync back
	crewState["context"] = "initial research data; added crewAI design"
	crewDataUpdated, err := json.Marshal(crewState)
	if err != nil {
		t.Fatalf("Failed to marshal updated CrewAI state: %v", err)
	}

	uCtxDataUpdated, err := crewBridge.ReadContext(crewDataUpdated)
	if err != nil {
		t.Fatalf("Failed to read updated CrewAI context: %v", err)
	}

	// 4. Hand off to AutoGen agent
	autoBridge, err := NewBridge("AutoGen")
	if err != nil {
		t.Fatalf("Failed to create AutoGen bridge: %v", err)
	}

	autoData, err := autoBridge.WriteContext(uCtxDataUpdated)
	if err != nil {
		t.Fatalf("Failed to write to AutoGen context: %v", err)
	}

	var autoState map[string]interface{}
	if err := json.Unmarshal(autoData, &autoState); err != nil {
		t.Fatalf("Failed to unmarshal AutoGen state: %v", err)
	}

	expectedAutoState := map[string]interface{}{
		"conversation_id": "session-001",
		"system_message":  "build_interop_bridge",
		"chat_history":    "initial research data; added crewAI design",
	}

	if !reflect.DeepEqual(autoState, expectedAutoState) {
		t.Errorf("AutoGen state mapping failed. Expected %v, got %v", expectedAutoState, autoState)
	}
}
