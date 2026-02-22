# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents evolve from single-instance entities to autonomous swarms (e.g., OpenClaw subagents), a significant challenge arises: **Context Fragmentation**. When a parent agent spawns a subagent, the subagent often lacks the high-level task context, user preferences, and security constraints established by the parent.

The Recursive Context Protocol (RCP) aims to solve this by standardizing how context is inherited and propagated across nested agent calls within the MCP Any bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize HTTP/JSON-RPC headers for context propagation.
    * Enable automatic inheritance of security scopes and session metadata.
    * Support "Context Trimming" to prevent token bloat in subagents.
* **Non-Goals:**
    * Implementing the subagent spawning logic itself (this remains the agent's responsibility).
    * Providing a persistent long-term memory (this is handled by the Shared KV Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator (e.g., OpenClaw)
* **Primary Goal:** Share secure context and task constraints between a parent "Project Manager" agent and 3 "Researcher" subagents without re-sending full environment variables.
* **The Happy Path (Tasks):**
    1. Parent agent calls an MCP tool to spawn a subagent, injecting an `X-MCP-Context-ID`.
    2. MCP Any core recognizes the Context ID and retrieves the associated parent state.
    3. Subagent makes a tool call; MCP Any automatically merges parent constraints (e.g., "Read-Only: true") into the subagent's request.
    4. Execution succeeds with consistent security and context.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Parent->>MCP_Any: Call Tool (Context: Session_A)
        MCP_Any-->>Parent: Result + Subagent_Token
        Parent->>Subagent: Spawn(Subagent_Token)
        Subagent->>MCP_Any: Call Tool (Auth: Subagent_Token)
        Note over MCP_Any: Resolve Parent(Session_A)
        MCP_Any->>Upstream: Exec Tool (Merged Context)
    ```
* **APIs / Interfaces:**
    * `X-MCP-Parent-ID`: Header for tracking lineage.
    * `X-MCP-Context-Scope`: Bitmask or JSON field for inherited permissions.
* **Data Storage/State:**
    * Transient context stored in MCP Any's internal session cache (TTL-bound).

## 5. Alternatives Considered
* **Implicit Context:** Guessing context based on IP or API Key. Rejected due to lack of precision in multi-tenant environments.
* **Full Context Passing:** Requiring agents to pass the entire context object in every call. Rejected as it causes "Context Bloat" and increases latency/token costs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Context inheritance must follow the principle of least privilege. A subagent cannot inherit *more* permissions than its parent.
* **Observability:** Traces must link parent and subagent tool calls using a unified `TraceID` derived from the `Parent-ID`.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
