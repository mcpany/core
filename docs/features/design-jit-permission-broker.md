# Design Doc: JIT Permission Broker
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agent swarms (e.g., OpenClaw, AutoGen) become more autonomous, they frequently encounter "Permission Deadlocks"—scenarios where an agent identifies a necessary tool for a task but lacks the pre-authorized capability token. Traditional static configuration requires human intervention to update the `mcp.yaml` or environment variables, which stalls autonomous workflows.

The JIT Permission Broker provides a protocol-native way for agents to request temporary, context-bound permission elevation. It acts as a middleman between the LLM and the Policy Engine, negotiating access based on task urgency, risk scoring, and human-in-the-loop (HITL) availability.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to request access to tools not in their initial bootstrap scope.
    * Support time-bound and session-bound "Leased Permissions."
    * Integrate automated risk scoring (e.g., "Is this a destructive operation?") to determine if HITL is required.
    * Provide a standardized JSON-RPC interface for permission requests.
* **Non-Goals:**
    * Automatically granting global/admin permissions.
    * Replacing the static Policy Firewall (Rego/CEL); instead, it generates temporary overrides for it.
    * Managing persistent user identities (handled by upstream auth).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous DevOps Agent
* **Primary Goal:** Execute a `database_migration` tool that was not initially authorized, to resolve a high-priority production incident.
* **The Happy Path (Tasks):**
    1. The Agent attempts to call `database_migration`.
    2. MCP Any rejects the call but includes a `capability_request_hint` in the error response.
    3. The Agent calls the `jit/request_permission` tool with the required scope and a justification (the incident report).
    4. The JIT Broker evaluates the risk and the Agent's current "trust score."
    5. The JIT Broker triggers a Slack/UI notification for human approval.
    6. The Human approves the request via the MCP Any Dashboard.
    7. The JIT Broker issues a temporary capability token valid for 30 minutes.
    8. The Agent retries the `database_migration` and succeeds.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Gateway: tools/call (database_migration)
        Gateway->>PolicyEngine: Is Authorized?
        PolicyEngine-->>Gateway: Denied (Missing Scope)
        Gateway-->>Agent: Error (403: Use jit/request_permission)
        Agent->>JITBroker: jit/request_permission(scope, intent)
        JITBroker->>RiskEngine: Analyze Request
        RiskEngine-->>JITBroker: Risk: High (Destructive)
        JITBroker->>HITLService: Create Approval Task
        HITLService-->>Human: UI/Webhook Notification
        Human->>HITLService: Approve (valid: 30m)
        HITLService-->>JITBroker: Approved
        JITBroker->>PolicyEngine: Inject Temporary Override
        JITBroker-->>Agent: Success (Scope Granted)
        Agent->>Gateway: tools/call (database_migration)
        Gateway-->>Agent: Result
    ```
* **APIs / Interfaces:**
    * `jit/request_permission`:
        * `scope`: string (e.g., `db:migrate:prod`)
        * `duration`: string (e.g., `30m`)
        * `justification`: string (Natural language intent)
    * `jit/status`: Check status of a pending request.
* **Data Storage/State:**
    * Temporary permissions are stored in the **Shared KV Store** (SQLite) with an expiration timestamp.
    * A background worker periodically prunes expired overrides.

## 5. Alternatives Considered
* **Always-on HITL**: Too much friction for low-risk tasks (e.g., read-only access).
* **Static Config Only**: Leads to "Permission Sprawl" where agents are over-provisioned "just in case."
* **LLM-Based Auto-Approval**: Rejected due to risk of prompt injection or agent hallucination bypass.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Every request must be tied to a verified `session_id`.
    * High-risk tools (FS write, DB delete) *must* require multi-factor human approval.
    * Prevent "Intent Spoofing" by validating the justification against recent trace logs.
* **Observability:**
    * Every JIT request and its outcome is logged in the Audit Trail.
    * Dashboard visualization for "Active Leases" and "Approval Bottlenecks."

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
