# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
In agent swarms, a parent agent often spawns subagents to perform specific tasks. Currently, each subagent must be initialized with its own full context, leading to "Token Bloat" and redundant configuration. The Recursive Context Protocol (RCP) allows subagents to inherit parent context, authentication, and state via standard MCP headers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce token usage in multi-agent handovers.
    *   Standardize context inheritance headers (e.g., `X-MCP-Parent-Context`).
    *   Allow subagents to "look up" parent state via the gateway.
*   **Non-Goals:**
    *   Persistent long-term memory (covered by Agent Blackboard).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars repeatedly.
*   **The Happy Path (Tasks):**
    1.  Parent Agent makes a tool call to spawn a subagent.
    2.  Parent Agent includes a `context_id` in the request metadata.
    3.  Subagent starts and queries MCP Any for `context_id`.
    4.  MCP Any provides the inherited context (e.g., specific FS paths allowed).
    5.  Subagent executes with limited, inherited scope.

## 4. Design & Architecture
*   **System Flow:**
    `Parent Agent -> [Context Headers] -> MCP Any -> Subagent`
*   **APIs / Interfaces:**
    *   New MCP Header: `X-MCP-Parent-Context-ID`.
    *   `tools/context/get`: Retrieve inherited context data.
*   **Data Storage/State:**
    In-memory session store or SQLite-backed "Context Registry".

## 5. Alternatives Considered
*   **Full Context Re-sending:** Rejected due to P95 token costs and latency.
*   **Shared Volume Mounts:** Rejected because it doesn't handle auth or non-file context.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Subagents can only inherit a subset of permissions (Scoped Inheritance).
*   **Observability:** Context lineage is tracked in Tracing to show parent-child relationships.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
