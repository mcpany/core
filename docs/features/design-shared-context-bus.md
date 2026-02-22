# Design Doc: Shared Context Bus
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agent swarms (e.g., CrewAI, AutoGen) become more prevalent, there is a critical need for a standardized way to share context between independent agents. Currently, context is either lost between tool calls or passed manually, which is inefficient and error-prone. MCP Any needs to provide a centralized "Bus" where agents can store and retrieve shared state.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a secure, ephemeral key-value store for agent context.
    *   Support scoped context (session-level, swarm-level).
    *   Enable real-time context updates via MCP notifications.
*   **Non-Goals:**
    *   Long-term persistent storage (this is for active session context).
    *   Complex relational data modeling.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
*   **The Happy Path (Tasks):**
    1.  Orchestrator initializes a `swarm_id` via MCP Any.
    2.  Agent A performs a research task and stores findings in the Shared Context Bus under the `swarm_id`.
    3.  Agent B retrieves the findings from the Bus using the same `swarm_id`.
    4.  Agent C updates the context with a final report.
    5.  MCP Any cleans up the context after the session expires.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>MCP Any: tools/call (context_store, {key: "findings", val: "..."})
        MCP Any->>ContextBus: Save to Memory/Redis
        Agent B->>MCP Any: tools/call (context_retrieve, {key: "findings"})
        ContextBus-->>MCP Any: Return "..."
        MCP Any-->>Agent B: Result
    ```
*   **APIs / Interfaces:**
    *   `context_store(key, value, scope)`
    *   `context_retrieve(key, scope)`
    *   `context_list(scope)`
*   **Data Storage/State:**
    *   In-memory store for local deployments.
    *   Redis/Etcd backend for distributed agent swarms.

## 5. Alternatives Considered
*   **Passing context in prompt:** Rejected due to context window limits and security risks (token leakage).
*   **Client-side state management:** Rejected because it requires all clients to implement the same logic, defeating the "Universal Gateway" purpose of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context is scoped to a specific `session_token` or `swarm_id`. Unauthorized agents cannot access other swarms' data.
*   **Observability:** All context operations are logged in the MCP Any audit trail.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
