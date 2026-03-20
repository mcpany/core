package interop

import (
	"encoding/json"
	"fmt"
)

// ContextBridge defines the universal bridge interface for mapping framework-specific state
// to a standardized Universal Agent Bus (UAB) context.
type ContextBridge interface {
	// ReadContext translates framework-specific state into the universal format.
	ReadContext(frameworkState []byte) ([]byte, error)
	// WriteContext translates universal context back into framework-specific state.
	WriteContext(universalContext []byte) ([]byte, error)
	// GetFrameworkName returns the name of the supported framework.
	GetFrameworkName() string
}

// UniversalContext represents the standardized context format.
type UniversalContext struct {
	SessionID string `json:"session_id"`
	Intent    string `json:"intent"`
	Memory    string `json:"memory"`
}

// openClawBridge implements ContextBridge for OpenClaw.
type openClawBridge struct{}

func (b *openClawBridge) GetFrameworkName() string { return "OpenClaw" }
func (b *openClawBridge) ReadContext(data []byte) ([]byte, error) {
	// Simulate mapping OpenClaw state
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	uctx := UniversalContext{
		SessionID: fmt.Sprintf("%v", state["oc_session"]),
		Intent:    fmt.Sprintf("%v", state["mission_root"]),
		Memory:    fmt.Sprintf("%v", state["shards"]),
	}
	return json.Marshal(uctx)
}
func (b *openClawBridge) WriteContext(data []byte) ([]byte, error) {
	var uctx UniversalContext
	if err := json.Unmarshal(data, &uctx); err != nil {
		return nil, err
	}
	state := map[string]interface{}{
		"oc_session":   uctx.SessionID,
		"mission_root": uctx.Intent,
		"shards":       uctx.Memory,
	}
	return json.Marshal(state)
}

// crewAIBridge implements ContextBridge for CrewAI.
type crewAIBridge struct{}

func (b *crewAIBridge) GetFrameworkName() string { return "CrewAI" }
func (b *crewAIBridge) ReadContext(data []byte) ([]byte, error) {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	uctx := UniversalContext{
		SessionID: fmt.Sprintf("%v", state["crew_id"]),
		Intent:    fmt.Sprintf("%v", state["task_goal"]),
		Memory:    fmt.Sprintf("%v", state["context"]),
	}
	return json.Marshal(uctx)
}
func (b *crewAIBridge) WriteContext(data []byte) ([]byte, error) {
	var uctx UniversalContext
	if err := json.Unmarshal(data, &uctx); err != nil {
		return nil, err
	}
	state := map[string]interface{}{
		"crew_id":   uctx.SessionID,
		"task_goal": uctx.Intent,
		"context":   uctx.Memory,
	}
	return json.Marshal(state)
}

// autoGenBridge implements ContextBridge for AutoGen.
type autoGenBridge struct{}

func (b *autoGenBridge) GetFrameworkName() string { return "AutoGen" }
func (b *autoGenBridge) ReadContext(data []byte) ([]byte, error) {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	uctx := UniversalContext{
		SessionID: fmt.Sprintf("%v", state["conversation_id"]),
		Intent:    fmt.Sprintf("%v", state["system_message"]),
		Memory:    fmt.Sprintf("%v", state["chat_history"]),
	}
	return json.Marshal(uctx)
}
func (b *autoGenBridge) WriteContext(data []byte) ([]byte, error) {
	var uctx UniversalContext
	if err := json.Unmarshal(data, &uctx); err != nil {
		return nil, err
	}
	state := map[string]interface{}{
		"conversation_id": uctx.SessionID,
		"system_message":  uctx.Intent,
		"chat_history":    uctx.Memory,
	}
	return json.Marshal(state)
}

// NewBridge returns a ContextBridge for the specified framework.
func NewBridge(framework string) (ContextBridge, error) {
	switch framework {
	case "OpenClaw":
		return &openClawBridge{}, nil
	case "CrewAI":
		return &crewAIBridge{}, nil
	case "AutoGen":
		return &autoGenBridge{}, nil
	default:
		return nil, fmt.Errorf("unsupported framework: %s", framework)
	}
}
