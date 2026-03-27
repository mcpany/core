# Design Doc: Dynamic Capability Leasing (DCL) Broker
**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
As agent swarms become more autonomous and long-lived, static capability models (where an agent has a set of permissions for its entire session) are increasingly inadequate. Today's research into Gemini CLI v0.41.0 highlights "Dynamic Capability Leasing" (DCL) as the solution. MCP Any needs a DCL Broker to ensure that tools are only accessible when an agent's active reasoning path explicitly proves a time-bound necessity.

## 2. Goals & Non-Goals
* **Goals:**
    * Bind tool permissions to the agent's verified reasoning lineage (RPW).
    * Implement ephemeral "Leases" that expire automatically upon task completion or intent shift.
    * Provide a Zero-Trust enforcement point within the Policy Firewall.
* **Non-Goals:**
    * Replacing the base RBAC/Policy system (DCL acts as an additional restrictive layer).
    * Managing the LLM's internal reasoning (MCP Any only validates the proof provided).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator.
* **Primary Goal:** Ensure a subagent can only access `fs:write` while it is actively performing the "Log Rotation" task, and loses that access immediately after.
* **The Happy Path (Tasks):**
    1. Parent agent delegates "Log Rotation" to Subagent B, including an RPW-watermarked reasoning branch.
    2. Subagent B requests a capability lease from the DCL Broker, providing the RPW proof.
    3. DCL Broker verifies the RPW lineage and grants a 60-second lease for `fs:write:/var/log`.
    4. Subagent B performs the rotation.
    5. Subagent B attempts to use `fs:write` for a different file 5 minutes later; the request is denied as the lease has expired.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
      A[Agent Request] --> B{DCL Broker}
      B -->|Verify RPW| C[Reasoning-Path Validator]
      C -->|Valid| D[Grant Ephemeral Lease]
      B -->|Check Lease| E{Policy Firewall}
      E -->|Authorized| F[Execute Tool]
      E -->|Expired/Invalid| G[Deny Access]
    ```
* **APIs / Interfaces:**
    * `LeaseRequest(AgentID, Capability, RPWProof) -> LeaseToken`
    * `ValidateLease(LeaseToken) -> Boolean`
* **Data Storage/State:**
    * In-memory TTL (Time-To-Live) cache for active leases, keyed by `AgentID` and `Capability`.

## 5. Alternatives Considered
* **Static Intent Scoping (UACO v1.0):** Rejected because it doesn't handle the temporal aspect of "Just-in-Time" agency; once an intent is granted, it remains valid for the session.
* **Manual HITL for every call:** Rejected due to "Approval Fatigue" and inability to scale with high-frequency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The DCL Broker must ensure that RPW proofs are cryptographically bound to the session to prevent replay attacks.
* **Observability:** All lease grants, expirations, and denials must be logged to the `Joint Quorum Attestation Ledger`.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
