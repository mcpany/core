# Design Doc: Deterministic Sub-agent Router

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the release of OpenClaw 2026.2.17, multi-agent swarms are becoming the default for complex tasks. However, these swarms often spawn sub-agents dynamically, leading to "identity drift" and "capability leakage." MCP Any needs a deterministic way to route tool calls and state between parents and children, ensuring that every sub-agent operates within a strictly defined, verifiable sandbox.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a "Sub-agent Registry" that tracks the lineage of agent sessions.
    *   Enforce "Capability Inheritance" where sub-agents can only access a subset of the parent's tools.
    *   Provide deterministic "Agent IDs" based on the task hash and parent identity.
    *   Mitigate "Rogue Sub-agent" risks by requiring parent attestation for every child tool call.
*   **Non-Goals:**
    *   Managing the LLM logic of the sub-agents themselves.
    *   Replacing framework-specific orchestration (e.g., OpenClaw's internal spawning).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Swarm Orchestrator.
*   **Primary Goal:** Spawn 5 sub-agents to analyze a codebase without giving them access to the `~/.ssh` directory or the parent's GitHub token.
*   **The Happy Path (Tasks):**
    1.  Parent agent requests a "Sub-agent Token" from MCP Any with a specific "Intent Scope."
    2.  MCP Any generates a unique, deterministic ID for the sub-agent.
    3.  Sub-agent attempts to call the `read_file` tool for `~/.ssh/id_rsa`.
    4.  The Router intercepts the call, verifies the sub-agent's restricted scope, and denies the request.
    5.  The Router logs the attempt and alerts the Parent agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Token Request**: Parent agent sends a `mcp_spawn_intent` message.
    - **Lineage Tracking**: The Router creates a node in the `Agent Lineage Graph`.
    - **Policy Enforcement**: Every tool call is wrapped in a "Lineage Context" that the Policy Firewall uses to make decisions.
*   **APIs / Interfaces:**
    - `mcp_session_spawn(parent_id, intent_scope) -> subagent_token`
    - `mcp_session_verify(token) -> identity_metadata`
*   **Data Storage/State:** Persistent lineage graph stored in the internal SQLite blackboard.

## 5. Alternatives Considered
*   **Static Scoping**: Giving all sub-agents the same permissions as the parent. *Rejected* due to high security risk.
*   **Manual Approval for Every Sub-agent**: *Rejected* as it breaks the autonomy of the swarm.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of "Zero Trust for Swarms." It ensures that even if one sub-agent is compromised, the blast radius is limited.
*   **Observability:** Visualizing the "Sub-agent Tree" in the UI is critical for debugging complex swarm behaviors.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
