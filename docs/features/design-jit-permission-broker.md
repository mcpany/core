# Design Doc: Just-In-Time (JIT) Permission Broker

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agents become more autonomous and perform long-running tasks, they often encounter "Permission Deadlocks"—situations where they need a specific tool or resource access that wasn't granted upfront. Static, "all-or-nothing" permission models are either too restrictive (stalling the agent) or too permissive (violating Zero Trust). The JIT Permission Broker solves this by providing a protocol for agents to request temporary, scoped permission escalations based on their current task intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable agents to request additional permissions (capabilities) dynamically during a session.
    *   Bind escalated permissions to a specific "Intent" and time-to-live (TTL).
    *   Provide an automated bridge to Human-In-The-Loop (HITL) approval flows for high-risk escalations.
    *   Maintain a tamper-proof audit log of all escalation requests and approvals.
*   **Non-Goals:**
    *   Automatically granting all requests (permissions must still be governed by policy).
    *   Permanent permission changes (escalations are session-bound or TTL-bound).

## 3. Critical User Journey (CUJ)
*   **User Persona:** DevOps Automation Agent.
*   **Primary Goal:** Safely escalate from `read-only` to `read-write` access on a specific production database to perform a hotfix, after human approval.
*   **The Happy Path (Tasks):**
    1.  The agent attempts a `db_write` tool call and is denied by the Policy Firewall.
    2.  The agent calls the `request_capability_escalation` tool, providing the required scope (`db:write:prod-cluster-1`) and the justification (the "Intent").
    3.  The JIT Permission Broker receives the request and triggers a HITL notification in the MCP Any UI.
    4.  The Human Admin reviews the intent and approves the temporary grant for 30 minutes.
    5.  The Broker updates the agent's session token with the new capability.
    6.  The agent retries the `db_write` call successfully.

## 4. Design & Architecture
*   **System Flow:**
    - **Request**: Agents use a standard MCP tool `mcpany_request_escalation`.
    - **Policy Check**: The `Policy Engine` (Rego/CEL) determines if the request is "Auto-Approvable," "Requires HITL," or "Hard Denied."
    - **Token Mutation**: Upon approval, the `Session Manager` injects a temporary capability claim into the agent's active session metadata.
    - **Enforcement**: The `Policy Firewall` middleware checks the mutated session token for every subsequent tool call.
*   **APIs / Interfaces:**
    - `tool: mcpany_request_escalation(capability: string, justification: string, duration_mins: int)`
    - `internal_api: /v1/escalations/approve` (called by UI)
*   **Data Storage/State:** Escalation state is tracked in the `Shared KV Store` and persisted in the Audit Logs.

## 5. Alternatives Considered
*   **Upfront Maximum Permissions**: Granting agents everything they *might* need. *Rejected* for violating the Principle of Least Privilege.
*   **Restarting Sessions with New Config**: Killing the agent and restarting it with a broader config. *Rejected* because it loses the agent's internal state and context.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Escalation requests themselves are a vector for "Agent Hijacking." Intents must be cross-verified against the session's recursive context lineage to ensure the request is consistent with the high-level goal.
*   **Observability:** The UI must clearly distinguish between "Base Permissions" and "JIT Escalations" in the security dashboard.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
