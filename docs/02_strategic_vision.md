# Strategic Vision: MCP Any

## Mission
MCP Any aims to be the universal infrastructure layer for AI agents, providing a configuration-driven gateway that bridges the gap between disparate APIs and the Model Context Protocol (MCP).

## Core Pillars
1. **Universal Adaptability**: Support for any protocol (REST, gRPC, CLI, etc.) through simple configuration.
2. **Security & Governance**: Zero Trust execution environment for autonomous agents.
3. **Observability**: Complete visibility into agent-tool interactions.

## Strategic Evolution: [2026-02-22]
### The Universal Agent Bus
With the explosive growth of local autonomous agents like **OpenClaw**, MCP Any's role shifts from a simple API gateway to a **Universal Agent Bus**.

### Key Strategic Patterns
*   **Zero Trust Governance**: As agents gain more autonomy (e.g., executing shell commands), MCP Any must enforce strict, Rego-based security policies at the tool-call level.
*   **Recursive Context Protocol**: Subagent swarms require a standardized way to inherit context and authentication across multiple levels of delegation. MCP Any will implement standardized headers for context inheritance.
*   **Shared State (Blackboard)**: Moving beyond stateless tool calls to provide a shared KV store (SQLite-backed) for agents to coordinate and persist memory safely.
