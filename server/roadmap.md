# Server Roadmap

## 1. Top Priorities: The Universal Agent Bus (New Strategic Focus)
*   **[Security] Policy Firewall Engine:** Implement Rego/CEL based hooking for tool calls.
*   **[Security] Granular Scopes:** implement capability-based token system (`fs:read:/tmp`).
*   **[Comms] Recursive Context Protocol:** Standardize headers for Subagent inheritance.
*   **[x] [State] Shared Key-Value Store:** Embedded SQLite "Blackboard" tool for agents.
*   **[x] [Security] HITL Middleware:** Suspension protocol for user approval flows.

## 2. Updated Roadmap

### Status: Active Development

#### Upcoming (2026-06-03 Evolution)
*   **[P0] Cross-Framework Attestation Translator (CFAT)**: Bridge proprietary TPM-bound reasoning paths (Gemini/Apple) to framework-neutral SRM formats. (Added: 2026-06-03)
*   **[P0] Atomic Shard Lock-Manager (ASLM)**: Kernel-level lock manager to prevent "Shard-Collision" corruption during granular context streaming. (Added: 2026-06-03)
*   **[P1] Shard Pre-fetching Engine**: Speculative loading of context fragments based on real-time intent analysis. (Added: 2026-06-03)
