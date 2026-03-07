# Design Doc: Universal State Handover (USH) Bridge

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
In complex agentic swarms (e.g., OpenClaw, CrewAI), tasks often move between specialized subagents. Currently, this "handoff" requires the parent agent to summarize the state, which is lossy and consumes context tokens. The **Universal State Handover (USH)** Bridge allows agents to pass their internal execution state (variables, local memory, and task progress) as a standardized, opaque reference or header. MCP Any acts as the secure registry and transport for these state objects.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a USH-compliant header protocol for MCP tool calls.
    *   Provide a secure repository (using the Shared KV Store) for storing agent state objects.
    *   Enable "Deep-Link" state references that allow Agent B to resume exactly where Agent A left off.
    *   Support "Lazy-State" where large objects are only fetched if needed.
*   **Non-Goals:**
    *   Translating internal variable formats between incompatible agent frameworks (frameworks must agree on the USH payload format).
    *   Persistent long-term storage of agent brains (focus is on active session handovers).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Systems Engineer.
*   **Primary Goal:** Seamlessly hand off a "Debug Task" from a Log-Analyzer Agent to a Fixer Agent without losing the identified variable traces.
*   **The Happy Path (Tasks):**
    1.  Log-Analyzer identifies a bug and its stack trace.
    2.  Log-Analyzer calls `ush_push_state(state_data={...trace...})` via MCP Any.
    3.  MCP Any returns a `ush_token`.
    4.  Log-Analyzer delegates to Fixer Agent, passing the `ush_token` in the task description.
    5.  Fixer Agent calls `ush_pull_state(token=ush_token)` to retrieve the full trace data.
    6.  Fixer Agent continues with the high-fidelity state without the parent agent needing to summarize the logs.

## 4. Design & Architecture
*   **System Flow:**
    - **State Ingestion**: Agents POST state blobs to `/ush/v1/state`.
    - **Reference Generation**: MCP Any generates a deterministic, session-bound UUID for the state.
    - **Header Injection**: MCP Any can automatically inject USH tokens into outgoing tool calls if configured.
*   **APIs / Interfaces:**
    - `ush_push_state(payload: Object) -> token: String`
    - `ush_pull_state(token: String) -> payload: Object`
    - `ush_list_states(session_id: String) -> tokens: List[String]`
*   **Data Storage/State:** State blobs are stored in the `Shared KV Store` (SQLite) with a configurable TTL (Time-To-Live) to prevent disk bloat.

## 5. Alternatives Considered
*   **Context Injection**: Just put the state in the LLM prompt. *Rejected* because it causes context bloat and increases token costs.
*   **Direct Agent-to-Agent Sockets**: *Rejected* because it bypasses security/audit logs and requires agents to manage network connections.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** USH tokens are session-bound. A token generated in Session A cannot be used in Session B. Payloads are encrypted at rest if they contain sensitive keys.
*   **Observability:** The UI will show "State Handoffs" in the session timeline, allowing developers to see exactly what data was passed between agents.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
