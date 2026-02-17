# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-17

## 1. Context and Scope
As AI agents evolve into complex multi-agent swarms (e.g., Claude Code subagents, OpenClaw), the need for a standardized way to pass context and security policies down the agent chain becomes critical. Currently, each subagent often starts with a "blank slate," leading to fragmented execution and security risks.

The Recursive Context Protocol (RCP) aims to standardize how MCP Any handles context inheritance, ensuring that every tool call in a chain is aware of its provenance and the constraints imposed by the parent agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize context propagation headers (e.g., `X-MCP-Agent-ID`, `X-MCP-Parent-Context`).
    * Enable subagents to automatically inherit security policies (Firewall rules) from their parents.
    * Provide a mechanism for "Context Compaction" to prevent token bloat during recursion.
* **Non-Goals:**
    * Implementing the agent orchestration logic itself (left to frameworks like CrewAI/OpenClaw).
    * Synchronizing full LLM state (memory) across agents (handled by the Shared KV Store feature).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Securely pass environment constraints (e.g., "Read-Only on /prod") from a Root Agent to 5 specialized subagents.
* **The Happy Path (Tasks):**
    1. Root Agent connects to MCP Any and establishes a session with a specific "Policy Profile".
    2. Root Agent spawns a subagent and passes an "Inheritance Token".
    3. Subagent connects to MCP Any using the token.
    4. MCP Any automatically applies the Root Agent's policies to the subagent's tool calls.
    5. MCP Any injects the parent's environmental context into the subagent's tool inputs where applicable.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        AgentA[Root Agent] -->|Tool Call + Context| MCPAny[MCP Any Gateway]
        MCPAny -->|Policy Applied| Upstream[Upstream Service]
        AgentA -->|Spawn| AgentB[Subagent]
        AgentB -->|Tool Call + ParentID| MCPAny
        MCPAny -->|Look up Parent Context| MCPAny
        MCPAny -->|Policy Merged| Upstream
    ```
* **APIs / Interfaces:**
    * New JSON-RPC notification: `notifications/context/push` to update active context.
    * Extended `initialize` params to include `parent_session_id`.
* **Data Storage/State:**
    * Transient session store in MCP Any linking subagents to parent sessions.

## 5. Alternatives Considered
* **Client-Side Propagation**: Forcing agents to manually pass context in every tool call. Rejected due to complexity and lack of security enforcement (agents could "forget" or bypass constraints).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Inheritance tokens must be short-lived and cryptographically bound to the parent session.
* **Observability**: Trace IDs must be propagated through the recursion to visualize the full "Call Tree" in the UI.

## 7. Evolutionary Changelog
* **2026-02-17:** Initial Document Creation.
