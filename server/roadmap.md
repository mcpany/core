# Server Roadmap

## 1. Top Priorities: The Universal Agent Bus (New Strategic Focus)
*   **[Security] Policy Firewall Engine:** Implement Rego/CEL based hooking for tool calls.
*   **[Security] Granular Scopes:** implement capability-based token system (`fs:read:/tmp`).
*   **[Comms] Recursive Context Protocol:** Standardize headers for Subagent inheritance.
*   **[State] Shared Key-Value Store:** Embedded SQLite "Blackboard" tool for agents.
*   **[Security] HITL Middleware:** Suspension protocol for user approval flows.

## 2. Updated Roadmap

### Status: Active Development

#### Upcoming (2026-05-06 Evolution)
*   **[P0] RAMS Temporal Memory Isolation**: Implement memory-mapped shard zeroing and rotation to mitigate SME. (Added: 2026-05-06)
*   **[P0] RBW Handshake Protocol**: Mandatory "Reasoning Proof" validation for all persistent WebSocket upgrades. (Added: 2026-05-06)
*   **[P1] JIT Capability Pruning (DCP)**: Middleware to dynamically modify tool schemas based on hardware-attested task state. (Added: 2026-05-06)

#### Upcoming (2026-02-23 Evolution)
*   **[P0] Recursive Context Protocol**: Finalize header-based context inheritance for swarms.
*   **[P0] Zero-Trust Subagent Scoping**: Implement intent-bound capability tokens.
*   **[P1] Environment Bridging Middleware**: Secure state sync between cloud sandboxes and local tools.
*   **[P1] Machine-Checkable Security Contracts**: Declarative tool safety models.
*   **[P0] Multi-Agent Session Management**: Session-aware middleware for agent coordination (Added: 2026-02-24).
*   **[P1] Unified MCP Discovery Service**: Automated registry for Stdio/HTTP/FastMCP servers (Added: 2026-02-24).

