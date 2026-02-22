# Design Doc: Standardized Recursive Context Protocol
**Status:** Draft
**Created:** 2025-05-22

## 1. Context and Scope
As agent swarms (like OpenClaw) become more hierarchical, "Manager" agents frequently call "Worker" subagents. These workers often need the same credentials, session IDs, or billing tags as the manager, but passing these through LLM prompts is insecure and inefficient. MCP Any needs a protocol-level way to propagate this context.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Standardize context propagation using `x-mcp-context-id` headers.
    *   Enable subagents to inherit parent context without LLM involvement.
    *   Support scoped credential injection based on context IDs.
*   **Non-Goals:**
    *   Implementing a full-blown identity provider (OIDC/SAML).
    *   Automatic translation between different context formats (beyond mapping).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
*   **The Happy Path (Tasks):**
    1.  Manager agent initiates a session with MCP Any, receiving a `context_id`.
    2.  Manager agent calls a subagent tool, passing the `context_id` in the metadata.
    3.  MCP Any intercepts the subagent call.
    4.  MCP Any looks up the parent context and injects required headers/auth into the subagent's upstream request.
    5.  Subagent executes successfully using the inherited credentials.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Tool Call (metadata: {context_id: "ABC"})
        MCP Any->>Context Registry: GetContext("ABC")
        Context Registry-->>MCP Any: {Headers: {"Authorization": "..."}, State: {...}}
        MCP Any->>Upstream: Request (with injected headers)
        Upstream-->>MCP Any: Response
        MCP Any-->>Agent: Tool Result
    ```
*   **APIs / Interfaces:**
    *   Internal `ContextRegistryInterface` for storing/retrieving session state.
    *   Middleware for header injection based on `context_id`.
*   **Data Storage/State:**
    *   Short-lived TTL-based storage in the "Blackboard" (SQLite).

## 5. Alternatives Considered
*   **Prompt-based passing**: Rejected due to security risks and token cost.
*   **Global Singleton Context**: Rejected because it doesn't support concurrent multi-tenant swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context IDs must be cryptographically signed or non-guessable (UUIDv4).
*   **Observability:** Context IDs should be logged in traces to allow full-swarm request tracking.

## 7. Evolutionary Changelog
*   **2025-05-22:** Initial Document Creation.
