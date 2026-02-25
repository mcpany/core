# Design Doc: JIT Permission Broker
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agent swarms (e.g., OpenClaw, CrewAI) grow more complex, they frequently encounter permission boundaries that require human intervention. This leads to "Approval Fatigue" for users and stalls autonomous workflows. The JIT (Just-In-Time) Permission Broker provides a mechanism for temporary, intent-based capability elevation, allowing agents to continue high-priority tasks under a "Conditional Trust" model.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to request temporary permission elevation for specific tool calls.
    * Use "Proof of Intent" (verifiable parent context/objective) to evaluate elevation requests.
    * Maintain a strict, immutable audit trail for all JIT elevations.
    * Support time-bound and call-count-bound permission leases.
* **Non-Goals:**
    * Replace the Policy Firewall (JIT works with it).
    * Provide permanent permission elevation.
    * Automate high-risk actions (e.g., wallet transfers) without ANY human oversight (still requires "Maximum Risk" overrides).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Swarm Orchestrator
* **Primary Goal:** Successfully complete a multi-step data migration that requires temporary sudo-level filesystem access without manual human approval for every sub-task.
* **The Happy Path (Tasks):**
    1. A subagent encounters an `EPERM` error while attempting to write to a protected directory.
    2. The subagent sends a `JITElevationRequest` to the MCP Any gateway, including its "Proof of Intent" (signed parent goal).
    3. The JIT Permission Broker validates the intent against the high-level project objective approved by the user.
    4. The Broker issues a 5-minute "Capability Lease" token.
    5. The subagent retries the tool call with the lease token; the Policy Firewall allows it.
    6. The task completes; the lease expires and is logged.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [JIT Request] -> JIT Broker -> [Validate Intent] -> [Issue Lease] -> Agent`
    `Agent -> [Tool Call + Lease] -> Policy Firewall -> [Check Lease] -> Upstream Tool`
* **APIs / Interfaces:**
    * `POST /v1/jit/elevate`: Accepts `intent_proof`, `requested_capabilities`, and `duration`. Returns a `lease_token`.
    * `GET /v1/jit/leases`: List active leases.
* **Data Storage/State:** Leases are stored in the shared SQLite state with an expiration TTL.

## 5. Alternatives Considered
* **Static Over-Provisioning**: Granting agents broad permissions upfront. Rejected due to Zero-Trust violations and high risk of "Agent Hijacking."
* **Always-on HITL**: Requiring human approval for every elevation. Rejected due to "Approval Fatigue" and workflow latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Proof of Intent" must be cryptographically signed by the orchestrator to prevent spoofing. Leases must be narrowly scoped to the requested tool and path.
* **Observability:** JIT elevations are flagged in the Audit Log with a special `JIT_ELEVATION` tag and links to the parent intent.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
