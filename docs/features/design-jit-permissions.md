# Design Doc: JIT Permission Broker (Lease-Based Auth)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
In complex agent swarms, pre-configuring all possible permissions (capabilities) is often impossible or leads to "over-provisioning," violating the principle of least privilege. Agents frequently encounter "Permission Deadlocks" where they need a specific tool access only for a subset of a task. The JIT Permission Broker allows agents to request temporary, task-scoped elevations (leases) that are dynamically evaluated by the Policy Engine and optionally approved by a human (via HITL).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement "Just-In-Time" capability escalation for MCP tools.
    *   Support time-bound and task-bound "Capability Leases."
    *   Integrate with the Policy Firewall for automated Rego-based evaluation of lease requests.
    *   Provide a standardized error code (`MCP_ERR_PERMISSION_DENIED_RETRYABLE`) that informs agents they can request JIT access.
*   **Non-Goals:**
    *   Permanent permission changes (all JIT elevations must expire).
    *   Bypassing the Policy Engine (JIT is an extension of the engine, not a replacement).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous Subagent (e.g., OpenClaw Specialist).
*   **Primary Goal:** Obtain temporary write access to a specific directory to fix a bug, without having permanent write access to the host.
*   **The Happy Path (Tasks):**
    1.  Subagent attempts to call `write_file` and receives a `RETRYABLE` permission error.
    2.  Subagent calls the `request_capability_lease` tool provided by the Broker, specifying the required scope (`fs:write:/src/bugfix/*`) and justification.
    3.  The JIT Broker evaluates the request against the Policy Engine (and triggers HITL if needed).
    4.  Broker issues a temporary JWT-backed capability token and injects it into the subagent's session.
    5.  Subagent retries the `write_file` call successfully.

## 4. Design & Architecture
*   **System Flow:**
    - **Lease Request**: Agents use a built-in `mcpany_request_lease` tool.
    - **Evaluation**: The Broker queries the `Policy Firewall` using the agent's "Intent Context" and "Lease Metadata."
    - **Token Injection**: Approved leases are stored in the `Shared KV Store` and automatically attached to subsequent tool calls within that session.
*   **APIs / Interfaces:**
    - `mcpany_request_lease(capability: string, duration: string, justification: string)`
    - Internal `LeaseManager` service in the Go server.
*   **Data Storage/State:** Leases are persisted in the SQLite "Blackboard" with an `expires_at` timestamp.

## 5. Alternatives Considered
*   **Static Permission Sets**: Manually defining every possible permission combination. *Rejected* due to lack of scalability and security risks.
*   **Always-HITL**: Requiring human approval for every new tool call. *Rejected* because it destroys agent autonomy and speed.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Prevent "Lease Squatting" where an agent requests broad permissions early and holds them. All leases must be as narrow as possible.
*   **Observability:** The "Supply Chain Attestation Viewer" in the UI will display active leases and their provenance (who requested, who approved).

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
