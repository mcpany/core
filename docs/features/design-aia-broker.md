# Design Doc: Active Intent Alignment (AIA) Broker
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms execute longer reasoning chains, "Intent Drift" becomes a significant risk. Specialist subagents may slowly deviate from the primary mission-root intent while maintaining cryptographically valid signatures. The AIA Broker provides a mechanism for periodic, hardware-attested "Alignment Heartbeats" to ensure the swarm remains semantically anchored to the original objective.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested semantic alignment verification.
    * Issue "Alignment Heartbeats" to specialist agents.
    * Detect and neutralize cumulative intent drift in deep reasoning chains.
    * Support cross-framework intent reconciliation.
* **Non-Goals:**
    * Automatically correcting reasoning (requires agent re-planning).
    * Storing full conversation histories (only semantic embeddings).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent System Architect
* **Primary Goal:** Ensure a Specialist Researcher agent doesn't "hallucinate" its way into an unrelated topic during a 2-hour autonomous session.
* **The Happy Path (Tasks):**
    1. The Mission-Root initializes the mission with the AIA Broker.
    2. Specialist agents receive "Alignment Tokens" bound to the mission-root.
    3. Periodically, the AIA Broker requests a reasoning fragment from the specialist.
    4. The AIA Broker compares the fragment's semantic embedding against the Mission-Root anchor.
    5. If alignment is within threshold, a new hardware-attested heartbeat is issued.
    6. If drift is detected, the broker revokes tool capabilities and triggers a parent re-planning cycle.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Anchor Intent| AIA[AIA Broker]
        SA[Specialist Agent] -->|Reasoning Trace| AIA
        AIA -->|Compare| Anchor[Mission Anchor]
        AIA -->|Attest| TPM[Hardware TPM]
        TPM -->|Valid| HB[Issue Heartbeat]
        TPM -->|Invalid| Revoke[Revoke Capabilities]
    ```
* **APIs / Interfaces:**
    * `POST /v1/alignment/anchor`: Establish the mission-root semantic anchor.
    * `POST /v1/alignment/verify`: Verify a reasoning fragment against the anchor.
    * `GET /v1/alignment/heartbeat`: Retrieve the latest attested heartbeat token.
* **Data Storage/State:**
    * Semantic anchors are stored as encrypted vectors in the Mission-Root Enclave.

## 5. Alternatives Considered
* **Continuous Full-Trace Review:** Rejected due to prohibitive token costs and latency.
* **Passive Metadata Monitoring:** Rejected as it cannot capture semantic nuances of "hallucinatory drift."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Heartbeats are hardware-bound and non-transferable. Revocation is propagated via the LFTA ARL.
* **Observability:** Real-time visualization of "Alignment Score" in the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
