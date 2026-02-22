# Design Doc: Recursive Context Protocol (RCP)
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents move from single-turn assistants to complex autonomous swarms (e.g., OpenClaw, Claude Code orchestrating subagents), there is a critical need for a standardized way to propagate context. Currently, spawning a subagent often results in "context amnesia" or requires insecurely passing raw environment variables. RCP aims to provide a secure, protocol-level mechanism for context inheritance within the MCP ecosystem.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable parent agents to pass scoped "Context Bundles" to subagents.
    *   Standardize RCP headers for inter-agent MCP communication.
    *   Support cryptographic signing of inherited context to prevent tampering.
    *   Allow subagents to contribute back to a shared parent state.
*   **Non-Goals:**
    *   Replacing the base MCP protocol (RCP is an extension/middleware).
    *   Managing the actual lifecycle of subagent processes (handled by the orchestrator).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator (e.g., OpenClaw user)
*   **Primary Goal:** Spawn a specialized "Security Research" subagent that inherits the parent's project context but has restricted access to production secrets.
*   **The Happy Path (Tasks):**
    1.  Parent agent initializes an RCP session with a scoped policy.
    2.  Parent agent calls `mcpany.spawn_subagent` with the RCP Context Bundle.
    3.  MCP Any validates the bundle and signs the subagent's session.
    4.  Subagent executes tools within the inherited constraints.
    5.  Subagent results are merged back into the parent's "Blackboard" state.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Parent Agent->>MCP Any: CallTool(spawn_subagent, context_bundle)
        MCP Any->>Policy Engine: Validate Scope
        Policy Engine-->>MCP Any: Approved (Signed)
        MCP Any->>Subagent: Initialize(mcp_config, rcp_headers)
        Subagent->>MCP Any: CallTool(restricted_tool)
        MCP Any->>Policy Engine: Verify RCP Signature
        Policy Engine-->>MCP Any: Allow
        MCP Any-->>Subagent: Result
    ```
*   **APIs / Interfaces:**
    *   `X-RCP-Bundle-ID`: Unique identifier for the inherited context.
    *   `X-RCP-Signature`: JWT or similar signed hash of the context and constraints.
    *   `mcpany.context.get()`: Tool for subagents to retrieve inherited state.
*   **Data Storage/State:**
    *   Temporary in-memory storage for active RCP sessions.
    *   Persistent SQLite "Blackboard" for long-running swarm state.

## 5. Alternatives Considered
*   **Raw Env Var Passing:** Rejected due to security risks (leakage) and lack of structure.
*   **State-in-Prompt:** Rejected due to "Context Bloat" and token costs; unreliable for complex data.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All RCP bundles must be signed by the MCP Any core. Subagents cannot escalate their own scopes.
*   **Observability:** RCP headers will be propagated to all audit logs, allowing for a "Trace Waterfall" of parent-child tool calls.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
