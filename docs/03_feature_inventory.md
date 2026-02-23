# Feature Inventory: MCP Any

This is the rolling masterlist of priority features, tracked by their status and strategic importance.

## Priority Features (P0)
- **Policy Firewall Engine:** Rego/CEL based hooking for tool calls. Ensuring Zero Trust.
- **HITL Middleware:** Suspension protocol for user approval flows.
- **Recursive Context Protocol (RCP):** Standardized context inheritance for subagents. (Added 2026-02-23)
- **Zero Trust Tool Sandboxing:** WASM/Docker-based isolation for tool execution. (Added 2026-02-23)

## High Priority Features (P1)
- **Shared Key-Value Store (Blackboard):** Embedded SQLite "Blackboard" tool for agents. (Added 2026-02-23)
- **Granular Scopes:** Capability-based token system (`fs:read:/tmp`).
- **Tool Playground & Explorer:** Auto-generated forms and history visualization.

## Medium Priority Features (P2)
- **Team Configuration Sync:** Securely sync configs and secrets across teams.
- **Smart Error Recovery:** Internal LLM loop for self-healing tool calls.
- **Live Marble Diagrams:** Reactive visualization of agent flows.

## Feature Grooming: [2026-02-23]
- **Added:** "Recursive Context Protocol (RCP)" to P0. Critical for solving swarm context loss.
- **Added:** "Zero Trust Tool Sandboxing" to P0. Essential response to OpenClaw security vulnerabilities.
- **Added:** "Shared Key-Value Store (Blackboard)" to P1. Needed for multi-agent state management.
- **Promoted:** "Policy Firewall Engine" remains P0, now explicitly tied to the Zero Trust strategic pillar.
