# Design Doc: JIT Capability Negotiator

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rise of "Shadow Tools" and complex multi-agent swarms, static permission models (e.g., permanent API keys with broad scopes) are no longer sufficient. Agents often require elevated privileges for short-lived, specific tasks (e.g., a "refinement agent" needing write access to a specific directory for 5 minutes). The JIT Capability Negotiator provides a standard way for agents to request, and users to grant, time-bound, intent-scoped capabilities.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a JSON-RPC based negotiation protocol for agents to request capabilities.
    *   Integrate with the HITL (Human-in-the-Loop) middleware for user approval of sensitive requests.
    *   Support cryptographic signatures for all capability tokens to ensure non-repudiation.
    *   Enforce TTL (Time-To-Live) on all granted capabilities.
*   **Non-Goals:**
    *   Replacing the base Policy Firewall (it works in conjunction with it).
    *   Managing the LLM's prompt itself (focus is on tool-access capabilities).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a multi-agent swarm for codebase refactoring.
*   **Primary Goal:** Grant a subagent temporary write access to a specific module without exposing the entire project.
*   **The Happy Path (Tasks):**
    1.  Parent agent spawns a "Refiner" subagent.
    2.  Subagent attempts to call `write_file` on `src/utils.py`.
    3.  MCP Any intercepts the call, identifying that the subagent lacks the permanent capability.
    4.  Subagent sends a `mcp_negotiate_capability` request with intent: "Refactor error handling in utils.py".
    5.  MCP Any triggers a HITL notification to the user's UI.
    6.  User approves the 10-minute "Scoped-Write" token.
    7.  Subagent proceeds with the task; token expires automatically after 10 minutes.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: Middleware catches `Unauthorized` errors from the Policy Firewall.
    - **Negotiation Trigger**: Returns a `403 Forbidden` with a `negotiation_required: true` hint.
    - **HITL Bridge**: Forwards the request to the `HITLMiddleware`.
    - **Token Issuance**: Generates a JWT signed by the MCP Any instance key, containing `scope`, `intent`, and `exp`.
*   **APIs / Interfaces:**
    - `mcp_negotiate_capability(request_id, scopes[], intent_description)`
    - `mcp_list_active_capabilities()`
*   **Data Storage/State:** In-memory store for active tokens, backed by SQLite for persistence across restarts.

## 5. Alternatives Considered
*   **Static Config Updates**: Forcing the user to edit `mcp.yaml` every time. *Rejected* - Too much friction for autonomous swarms.
*   **Broad "Sudo" Mode**: Giving agents full access for a session. *Rejected* - Violates principle of least privilege.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** JIT is the "Just-in-Time" pillar of Zero Trust. It minimizes the "Blast Radius" of a compromised subagent.
*   **Observability:** Capability requests and grants are logged in the Audit Log with the associated `intent_description`.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
