# Interop Blueprint: Universal Agent Bus (2026)

## 1. Overview
As part of our mission to evolve the MCP Any "Universal Agent Bus," we have successfully implemented an interoperability layer designed to support major AI standards. This blueprint outlines the specific framework adapters we have built to allow seamless state and context synchronization across heterogeneous swarms.

## 2. Supported Frameworks
The Interop Bridge now officially supports the following Multi-Agent frameworks by translating their proprietary context states into the `UniversalContext` standard:

*   **OpenClaw**: Connects with OpenClaw's context engines.
    *   State Mapping: `oc_session` $\rightarrow$ SessionID, `mission_root` $\rightarrow$ Intent, `shards` $\rightarrow$ Memory.
*   **CrewAI**: Supports CrewAI's task-driven crew orchestrations.
    *   State Mapping: `crew_id` $\rightarrow$ SessionID, `task_goal` $\rightarrow$ Intent, `context` $\rightarrow$ Memory.
*   **AutoGen**: Bridges with Microsoft AutoGen's conversational agents.
    *   State Mapping: `conversation_id` $\rightarrow$ SessionID, `system_message` $\rightarrow$ Intent, `chat_history` $\rightarrow$ Memory.

## 3. The `ContextBridge` Pattern
The core of our interoperability relies on the `ContextBridge` interface. Each framework implements its own bridge adapter within the `src/interop/` package.

### 3.1 `UniversalContext`
We use a centralized structural representation of agent context to standardize handoffs:
```go
type UniversalContext struct {
	SessionID string `json:"session_id"`
	Intent    string `json:"intent"`
	Memory    string `json:"memory"`
}
```

### 3.2 Bridging API
The API provides two primary operations for every supported framework:
1.  **`ReadContext(frameworkState []byte)`**: Transforms a framework's native context format into a serialized `UniversalContext`.
2.  **`WriteContext(universalContext []byte)`**: Transforms a serialized `UniversalContext` back into the framework's native context format.

## 4. Swarm Simulations & Validation
Our continuous integration relies on Bazel targets (e.g., `//src/interop/...`). The integration tests (`swarm_test.go`) explicitly simulate multi-agent swarms by passing the context sequentially from OpenClaw $\rightarrow$ CrewAI $\rightarrow$ AutoGen. The tests verify that the "Intent" (task goals) and "Memory" (shards/context/chat history) are preserved and faithfully mapped.
