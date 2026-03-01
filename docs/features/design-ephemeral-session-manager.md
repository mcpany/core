# Design Doc: Ephemeral Session Manager

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents move towards "one-off" ephemeral tasks (e.g., a single PR review via Claude Code or a quick search via Gemini CLI), the overhead of session initialization and the risk of state leakage have become primary concerns. Current systems either retain too much state (increasing security risk) or take too long to "warm up" (hurting UX). MCP Any needs a dedicated manager to handle these short-lived agent lifecycles securely and efficiently.

## 2. Goals & Non-Goals
*   **Goals:**
    *   **Predictive Context Warmup**: Pre-load tool schemas and relevant state based on the initial prompt's intent.
    *   **Secure Session Isolation**: Ensure that context from one ephemeral session never leaks into another.
    *   **Automated State Purge**: Wipe all session-bound KV pairs and tool execution logs immediately upon agent disconnection.
    *   **Low-Latency Initialization**: Reduce "First Call" latency to sub-200ms for ephemeral swarms.
*   **Non-Goals:**
    *   Replacing long-term memory systems (this is for *ephemeral* state).
    *   Managing the lifecycle of the agent process itself (MCP Any focuses on the communication and state layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer running a one-time `claude-code` command to fix a bug.
*   **Primary Goal:** Execute the task with zero persistent trace of the sensitive data processed during the session.
*   **The Happy Path (Tasks):**
    1.  User starts an ephemeral agent session.
    2.  MCP Any detects the new session ID and begins "Predictive Warmup" based on the first few tokens of the prompt.
    3.  Agent calls tools; state is stored in a `session_local` KV store.
    4.  Task completes; agent disconnects.
    5.  MCP Any triggers the "Secure Purge" routine, wiping all `session_local` data and audit logs for that session ID.

## 4. Design & Architecture
*   **System Flow:**
    - **Session Tracker**: Hooks into the transport layer (Stdio/WS/Pipe) to detect session start/end.
    - **Warmup Engine**: A small, fast LLM or heuristic model that maps prompts to tool "clusters" and pre-fetches them.
    - **Isolated KV Store**: A partitioned SQLite database or in-memory map scoped by `SessionID`.
*   **APIs / Interfaces:**
    - `POST /v1/session/init { intent_hint: string }` -> Returns `SessionID` and pre-fetched tools.
    - `DELETE /v1/session/{id}` -> Manually trigger a purge.
*   **Data Storage/State:**
    - Temporary, in-memory state for tool results.
    - Session-scoped partitions in the Shared KV Store.

## 5. Alternatives Considered
*   **Stateless Execution**: Forcing agents to be entirely stateless. *Rejected* as complex tasks (multi-step tool calls) require intermediate state.
*   **External Cleanup Scripts**: Relying on cron jobs or external scripts to clean up logs. *Rejected* as it leaves a window for state leakage.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The "Secure Purge" is a critical component of the Zero Trust posture for ephemeral agents.
*   **Observability:** Provide "Purge Confirmation" logs to the system administrator to verify state was wiped.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
