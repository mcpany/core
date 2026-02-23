# Design Doc: Recursive Context Protocol (RCP)
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As AI agents move from single-task execution to complex swarms, the loss of context during subagent delegation has become a major bottleneck. Currently, each subagent call starts with a fresh or manually copied context, leading to hallucinations or repetitive work. MCP Any needs to provide a standardized way for context to flow recursively through agent chains.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Define a standardized set of headers/metadata for context inheritance.
    *   Enable automatic propagation of "Blackboard" state to subagents.
    *   Provide a "Parent-Child" relationship mapping for tool calls.
*   **Non-Goals:**
    *   Implementing the LLM-side context compression logic.
    *   Replacing the core MCP protocol (RCP will be built on top of it).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
*   **The Happy Path (Tasks):**
    1. Orchestrator initializes a session with `X-MCP-Context-ID`.
    2. Orchestrator calls Tool A, which spawns Subagent B.
    3. Tool call to Subagent B automatically includes `X-MCP-Parent-Context`.
    4. Subagent B accesses the shared Blackboard via MCP Any using the inherited context.
    5. MCP Any enforces that Subagent B only sees state relevant to the parent's scope.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Orchestrator->>MCP Any: ToolCall (headers: Context-ID=123)
        MCP Any->>Tool A: Execute
        Tool A->>MCP Any: ToolCall (Subagent B, headers: Parent-ID=123)
        MCP Any->>Subagent B: Execute with Context 123
    ```
*   **APIs / Interfaces:**
    *   New Header: `X-MCP-Context-ID`: Unique session identifier.
    *   New Header: `X-MCP-Parent-ID`: Pointer to the caller agent's context.
    *   `Blackboard` Tool: `get_state(key)`, `set_state(key, value)`.
*   **Data Storage/State:**
    *   In-memory TTL cache for session context mapping.
    *   Persistent SQLite storage for long-running "Blackboard" state.

## 5. Alternatives Considered
*   **Manual Context Passing**: Rejected because it increases prompt overhead and is prone to developer error.
*   **Global Shared State**: Rejected due to security concerns (any agent could see any other agent's secrets).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context inheritance must be explicitly authorized. Subagents cannot "escalate" context access without parent permission.
*   **Observability:** All inherited calls must be tagged in the "Marble Diagram" for visualization of the dependency tree.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
