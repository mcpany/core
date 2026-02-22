# Strategic Vision: MCP Any (The Universal Agent Bus)

## Core Mission
To provide the indispensable infrastructure layer for the AI Agent era, enabling seamless, secure, and observable tool integration across all agentic frameworks.

## Strategic Evolution: [2025-05-22]

### 1. Standardized Recursive Context Inheritance
Today's research into OpenClaw and hierarchical swarms reveals a critical gap in how subagents inherit security context and session state.
*   **Pattern Match**: We can evolve the MCP protocol by introducing a standardized `x-mcp-context-id` header that MCP Any tracks.
*   **Strategic Move**: Implement a "Context Registry" within MCP Any that allows subagents to fetch scoped credentials without them ever touching the LLM prompt.

### 2. Zero Trust Tool execution (Rego-based)
With the rise of "Agent Hijacking" via tool outputs, a simple static policy is no longer sufficient.
*   **Pattern Match**: Borrowing from Kubernetes security, we will integrate Open Policy Agent (OPA) / Rego to allow fine-grained, dynamic tool-call validation.
*   **Strategic Move**: MCP Any will serve as the "Enforcement Point" for all agent actions, blocking suspicious tool sequences (e.g., `ls` followed by `rm` on the same path) unless HITL approval is granted.

### 3. The "Shared Blackboard" Pattern
Inter-agent communication currently lacks a reliable, shared state mechanism.
*   **Pattern Match**: Agents need a "Sidecar Memory".
*   **Strategic Move**: Expose a built-in `blackboard` tool in MCP Any that provides a scoped KV store with locking, allowing multiple agents to synchronize their world state safely.
