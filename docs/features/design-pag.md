# Design Doc: Probabilistic Attestation Gate (PAG)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
In high-density agent swarms, continuous hardware-bound attestation (TPM signatures) for every inter-agent coordination fragment adds significant latency (often 50ms-100ms per call). This leads to "Attestation Fatigue" and "Cognitive Stall" where the swarm spends more time verifying identity than reasoning.

The Probabilistic Attestation Gate (PAG) introduces a risk-aware, non-deterministic verification model. It allows the infrastructure to dynamically scale the attestation frequency based on real-time risk scores, providing high performance for routine coordination while maintaining absolute sovereignty for high-stakes tool calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a dynamic "Risk Scoring" engine that evaluates tool impact and reasoning confidence.
    * Support "Lightweight Session Heartbeats" for low-risk attestation.
    * Trigger full TPM-bound attestation automatically when risk thresholds are exceeded.
    * Achieve a 70% reduction in coordination latency for low-stakes swarm interactions.
* **Non-Goals:**
    * Replacing hardware attestation entirely (PAG is a frequency-management layer, not a replacement for TPM).
    * Providing security for external (non-mesh) tool calls.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Engineer
* **Primary Goal:** Reduce the "Coordination Tax" for a 50-agent parallel research swarm without compromising mission-root integrity.
* **The Happy Path (Tasks):**
    1. The engineer enables PAG on the mesh gateway with a "Balanced" risk policy.
    2. A subagent performs 10 consecutive "Read-Only" Blackboard lookups.
    3. PAG assigns these a low risk score and verifies them via lightweight heartbeats (sub-1ms overhead).
    4. The subagent then attempts a `git push` command (high-risk).
    5. PAG detects the risk spike and forcefully triggers a full hardware-attested handshake.
    6. The mission continues with 90% of coordination happening at "mesh-speed" while high-risk actions remain TPM-locked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Request] --> PAG[PAG Middleware]
        PAG --> Risk[Risk Engine]
        Risk -->|Low Risk| Cache[Verify Session Heartbeat]
        Risk -->|High Risk| TPM[Trigger TPM Signature]
        Cache --> Backend[Process Request]
        TPM --> Backend
    ```
* **APIs / Interfaces:**
    * `x-mcp-attestation-mode`: Header to signal `probabilistic` or `mandatory` modes.
    * `GET /pag/risk-scores`: Telemetry endpoint for real-time risk analysis.
* **Data Storage/State:**
    * Risk profiles are stored in the Mesh-Resident Attestation Registry.

## 5. Alternatives Considered
* **Adaptive Jitter:** Rejected as a primary solution because it focuses on side-channel defense rather than attestation overhead.
* **Trust Leases (LFTA):** PAG evolves LFTA by adding non-deterministic, risk-aware triggers instead of simple time-bound leases.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The PAG must include "Anti-Gaming" logic to prevent subagents from "warming up" low-risk scores before an attack.
* **Observability:** Attestation "Skips" are logged in the audit trail with the corresponding risk justification.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
