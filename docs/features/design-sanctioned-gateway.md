# Design Doc: Sanctioned Gateway Audit & Policy Layer
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The rapid adoption of autonomous agents like OpenClaw has created a governance vacuum. Enterprise users are deploying agents that can call thousands of tools without oversight, leading to the "Shadow Agent" problem. MCP Any must evolve into a "Sanctioned Gateway"—a mandatory, high-performance interception layer that enforces security policies and maintains a complete audit trail for every tool interaction.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept every tool call and response across all MCP transports.
    * Enforce granular, identity-based access control policies (Rego/CEL).
    * Generate cryptographically signed audit logs for SOC2 compliance.
    * Provide real-time policy evaluation with <5ms latency overhead.
* **Non-Goals:**
    * Implementing the tools themselves (gateway only).
    * Real-time content filtering (this is handled by the model-level guardrails).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator.
* **Primary Goal:** Ensure that no agent can call the "Delete Database" tool without explicit human approval or a high-confidence "Safe Intent" signal.
* **The Happy Path (Tasks):**
    1. Administrator defines a global policy: `deny if tool.name == "delete_db" and not request.human_approved`.
    2. An OpenClaw agent attempts to call `delete_db`.
    3. The `SanctionedGatewayMiddleware` intercepts the call and evaluates the Rego policy.
    4. The call is suspended, and a notification is sent to the HITL (Human-in-the-Loop) dashboard.
    5. After approval, the gateway signs the audit record and allows the call to proceed.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: All inbound JSON-RPC `tools/call` requests are routed through the `PolicyEngine`.
    - **Evaluation**: The `PolicyEngine` matches the request against pre-compiled Rego/CEL rules.
    - **Audit Logging**: Successful and denied calls are streamed to a persistent, immutable log (e.g., BoltDB or a remote SIEM).
* **APIs / Interfaces:**
    - `POST /v1/policy/evaluate`: Internal endpoint for pre-call verification.
    - `GET /v1/audit/stream`: Real-time audit log stream for the UI.
* **Data Storage/State:** Policies are stored in the `Shared KV Store`. Audit logs are stored in a write-ahead log (WAL) for durability.

## 5. Alternatives Considered
* **Agent-Side Enforcement**: Relying on agent frameworks to enforce policies. *Rejected* because it is bypassable by rogue agents.
* **Tool-Side Enforcement**: Adding security logic to every MCP server. *Rejected* because it is unscalable and lacks centralized governance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Gateway itself must be the "Root of Trust." All policy changes require MFA-Attestation.
* **Observability:** Integration with the "Audit & Security Dashboard" for real-time visualization of blocked calls.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
