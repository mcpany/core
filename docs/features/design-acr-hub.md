# Design Doc: Autonomous Capability Revocation (ACR) Hub
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms become more autonomous and specialized, the risk of "Intent
Drift" increases. Specialist agents may slowly deviate from the primary mission
root while maintaining valid cryptographic signatures, creating a "Drift Window"
where they can still execute authorized tools despite being semantically
misaligned.

The Autonomous Capability Revocation (ACR) Hub solves this by integrating
directly with Active Intent Alignment (AIA) heartbeats. It provides a mechanism
for sub-millisecond, autonomous revocation of agent capabilities across the
mission scope when misalignment is detected, ensuring that security is
reactively enforced without waiting for human intervention.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond capability revocation triggered by AIA drift
      signals.
    * Provide a centralized registry for session-bound agent capabilities.
    * Ensure revocation propagates across heterogeneous framework boundaries
      (OpenClaw, Claude Code, Gemini CLI).
    * Mandate hardware-attested re-authorization to restore revoked
      capabilities.
* **Non-Goals:**
    * Replacing the primary Policy Engine (ACR acts as a reactive override).
    * Managing long-term persistent permissions (ACR focuses on active session
      leases).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Automatically neutralize a compromised or drifting subagent
  before it can exfiltrate data via authorized tools.
* **The Happy Path (Tasks):**
    1. A specialist agent starts a high-trust reasoning task with
       hardware-attested capabilities.
    2. The AIA Broker monitors the agent's reasoning trace and detects semantic
       drift from the mission root.
    3. The AIA Broker issues a high-priority ACR signal to the Hub.
    4. The ACR Hub immediately invalidates all hardware-attested capability
       tokens for that agent's session ID.
    5. The next tool call attempted by the agent is rejected by the Gateway due
       to revoked attestation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>AIA Broker: Reasoning Trace (Heartbeat)
        AIA Broker->>AIA Broker: Semantic Alignment Check
        AIA Broker-->>ACR Hub: DRIFT_DETECTED (SessionID)
        ACR Hub->>Capability Registry: Invalidate(SessionID)
        Agent->>Gateway: Tool Call (Attested Token)
        Gateway->>ACR Hub: Validate(Token)
        ACR Hub-->>Gateway: REVOKED
        Gateway-->>Agent: 403 Forbidden (Alignment Failure)
    ```
* **APIs / Interfaces:**
    * `POST /v1/acr/revoke`: Internal endpoint for AIA Broker to signal
      revocation.
    * `GET /v1/acr/check`: Gateway endpoint to verify token validity against
      the revocation list.
* **Data Storage/State:**
    * High-speed, in-memory Bloom filter for active revocation list (ARL) to
      ensure sub-millisecond lookups.
    * Persistent audit log in SQLite for forensic analysis of revocation events.

## 5. Alternatives Considered
* **Manual Revocation:** Rejected due to the "Machine-Speed" nature of agentic
  attacks; human latency is too high.
* **Periodic Re-Attestation:** Rejected as a primary mechanism because it still
  leaves a window of vulnerability between attestation cycles. ACR provides
  immediate response.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ACR Hub itself must be protected by
  hardware-attested identity to prevent "Denial of Capability" attacks by
  compromised agents attempting to spoof revocation signals.
* **Observability:** Revocation events are streamed to the "Local Security
  Violation Monitor" with full reasoning trace context leading to the drift.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
