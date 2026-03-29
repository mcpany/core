# MCP Any - Interop Blueprint

## Supported AI Standards and Frameworks

This document serves as the architectural blueprint detailing the agent frameworks and communication protocols currently integrated and supported by the universal adapter bus.

### 1. OpenClaw
- **Focus**: Adaptive reasoning and context synchronization.
- **Supported Capabilities**:
  - `adaptive_reasoning`: High-entropy data digestion.
  - `context_sync`: State-aligned multimodal context ingestion.

### 2. CrewAI
- **Focus**: Task delegation and role-based agent management.
- **Supported Capabilities**:
  - `task_delegation`: Routing tasks to specialized roles (e.g., data analyst, generalist).
  - `role_discovery`: Dynamic role registration and auth token management.

### 3. AutoGen
- **Focus**: Multi-agent chat and stateful subagent execution.
- **Supported Capabilities**:
  - `multi_agent_chat`: Conversational sub-agent executions.
  - `subagent_exec`: Compiling and running complex, step-by-step tasks with persistent history checkpoints.

### 4. ACP (Agent Communication Protocol)
- **Focus**: Standardized inter-agent messaging and discovery.
- **Supported Capabilities**:
  - `agent_messaging`: Sending intent-driven messages across the swarm.
  - `capability_discovery`: Identifying features exposed by neighboring frameworks.

---
*All integrated frameworks support hardware-attested multimodal memory shards (UMMB).*
