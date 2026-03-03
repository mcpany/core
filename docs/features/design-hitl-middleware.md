# Design Doc: Human-in-the-Loop (HITL) Middleware
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As AI agents move toward higher autonomy and swarm-based execution, the risk of "runaway" actions (expensive tool calls, destructive file edits, or unauthorized data exfiltration) increases. Users are currently faced with a binary choice: full manual approval (high friction) or full autonomy (high risk).

MCP Any needs a sophisticated HITL Middleware that provides "Delegate-able Approvals"—allowing users to set time-bound, budget-bound, or scope-bound permissions that agents can use autonomously until a threshold is reached.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized suspension protocol for any MCP tool call.
    * Support asynchronous approval flows (the agent waits without blocking the server).
    * Implement "Approval Delegation" (e.g., "Allow this agent to spend up to $5.00 on tools without asking").
    * Integrate with the UI for real-time notifications.
* **Non-Goals:**
    * Implementing the LLM logic for *deciding* what needs approval (this is handled by the Policy Firewall).
    * Managing user authentication (handled by the Identity Mesh).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer using an autonomous swarm for infrastructure migration.
* **Primary Goal:** Approve a batch of "Terraform Apply" calls while delegating minor "Read" actions.
* **The Happy Path (Tasks):**
    1. The Engineer configures a "Session Budget" of $10.00 and allows all `fs:read` calls.
    2. The Swarm initiates 50 `fs:read` calls; MCP Any auto-approves them.
    3. The Swarm initiates a `terraform:apply` call.
    4. MCP Any detects a "High Risk" tool and suspends the execution.
    5. MCP Any sends a notification to the Engineer's UI with the tool arguments and a "Dry Run" preview.
    6. The Engineer clicks "Approve" in the UI.
    7. MCP Any resumes the tool call and returns the result to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Policy Firewall] -> [HITL Middleware] -> [Target Tool]`
    1. **Suspension**: Middleware intercepts the JSON-RPC request and stores it in a `PendingActions` table (SQLite).
    2. **Notification**: Emits a WebSocket event `hitl.pending` to the UI.
    3. **Callback**: The UI calls `/api/v1/hitl/approve/{id}` or `/api/v1/hitl/reject/{id}`.
    4. **Resumption**: Middleware retrieves the pending request and forwards it to the upstream MCP server.

* **APIs / Interfaces:**
    * `POST /api/v1/hitl/delegate`: Create an approval delegation (scope, budget, expiry).
    * `GET /api/v1/hitl/pending`: List all actions awaiting approval.
    * `POST /api/v1/hitl/approve/{id}`: Approve a specific action.

* **Data Storage/State:**
    Uses the embedded SQLite "Blackboard" to store `pending_actions` and `active_delegations`.

## 5. Alternatives Considered
* **Agent-Side HITL**: Rejected because it relies on the agent being "honest" about asking for permission. Middleware-side enforcement is required for Zero Trust.
* **Blocking Requests**: Rejected because it would lead to timeout issues in HTTP-based agents. Asynchronous "Suspension & Resume" is more robust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval tokens must be session-bound. Approving one tool call does not grant a blanket "Authorized" status for the next 10 minutes unless a Delegation is explicitly created.
* **Observability:** All HITL actions (Suspend, Approve, Reject, Auto-Approve) are logged in the Audit Log with the associated User ID.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
