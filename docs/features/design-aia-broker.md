# Design Doc: Active Intent Alignment (AIA) Broker
**Status:** Draft
**Created:** 2026-06-17

## 1. Context and Scope
As agent swarms become more complex and multi-layered, the risk of "Semantic
Drift" increases. Specialist agents, while remaining cryptographically valid,
may slowly diverge from the primary mission intent during long reasoning loops.
This "Intent Drift" can lead to unauthorized actions or inefficient resource
consumption.

The Active Intent Alignment (AIA) Broker acts as the authoritative host for
hardware-attested "Alignment Heartbeats." It periodically verifies that
specialist agent reasoning traces remain semantically aligned with the
mission-root intent, neutralizing cumulative drift and providing a foundation
for autonomous capability revocation.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue and verify hardware-attested "Alignment Heartbeats" for specialist
      agents.
    * Perform real-time semantic comparison between subagent reasoning and the
      mission-root manifest.
    * Provide a standardized interface for agents to report reasoning progress.
    * Trigger autonomous capability revocation (ACR) upon detection of
      significant alignment failure.
* **Non-Goals:**
    * Managing the primary agent execution loop (AIA is a monitoring and
      governance layer).
    * Enforcing tool-specific policies (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Monitor
* **Primary Goal:** Ensure that a deep chain of subagents remains focused on
  the user's original objective without manual oversight.
* **The Happy Path (Tasks):**
    1. A mission-root intent is established and cryptographically signed.
    2. Specialist subagents are spawned with session-bound AIA requirements.
    3. Periodically (e.g., every 5 reasoning steps), subagents submit a
       "Reasoning Heartbeat" to the AIA Broker.
    4. The AIA Broker uses the "Semantic Integrity Bridge" to compare the
       heartbeat against the mission-root.
    5. If aligned, the Broker issues a new hardware-attested "Alignment Token"
       allowing the session to continue.
    6. If drift is detected, the session is flagged for revocation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Subagent->>AIA Broker: Reasoning Heartbeat (Trace + Signatures)
        AIA Broker->>Mission Manifest: Retrieve Root Intent
        AIA Broker->>Semantic Engine: Compare(Heartbeat, RootIntent)
        Semantic Engine-->>AIA Broker: Alignment Score
        AIA Broker->>Subagent: New Alignment Token (if score > threshold)
    ```
* **APIs / Interfaces:**
    * `POST /v1/aia/heartbeat`: Endpoint for agents to submit reasoning
      traces.
    * `GET /v1/aia/status`: Endpoint to check the alignment health of a
      mission branch.
* **Data Storage/State:**
    * Heartbeat history is stored in an embedded SQLite database for
      auditability.
    * Active alignment tokens are cached in-memory with hardware-bound TTLs.

## 5. Alternatives Considered
* **Static Intent Check at Boot:** Rejected because it doesn't account for
  drift during the execution phase.
* **Parent-Only Monitoring:** Rejected because it creates a "Supervisor
  Bottleneck" and doesn't provide cross-framework alignment consistency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Heartbeats must be hardware-attested
  (TPM/Secure Enclave) to prevent "Spoofed Alignment" by a compromised agent.
* **Observability:** Alignment scores are visualized in the "Active Intent
  Alignment Monitor" UI.

## 7. Evolutionary Changelog
* **2026-06-17:** Initial Document Creation.
* **2026-06-18:** Integrated with ACR Hub for autonomous revocation.
