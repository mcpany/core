# Design Doc: Active Intent Alignment (AIA) Broker

**Status:** Draft
**Created:** 2026-03-19

## 1. Context and Scope

As AI agent swarms become more autonomous and specialized, the risk of "Intent Drift" increases. Specialist agents, while cryptographically verified, may begin to diverge from the mission-root intent during long-running reasoning chains. The AIA Broker provides a mechanism for continuous semantic verification of agent reasoning traces against the primary mission intent.

## 2. Goals & Non-Goals

* **Goals:**
    * Implement a hardware-attested "Alignment Heartbeat" for specialist agents.
    * Provide real-time semantic comparison between subagent traces and mission-root manifests.
    * Enable automated interdiction/revocation of capabilities upon detected drift.
* **Non-Goals:**
    * It will not attempt to rewrite agent reasoning traces.
    * It will not manage agent resource budgets (handled by RBF).

## 3. Critical User Journey (CUJ)

* **User Persona:** Security-conscious Enterprise Swarm Orchestrator
* **Primary Goal:** Ensure that a specialized "Code Auditor" subagent does not diverge into "Code Generation" during a security review mission.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns a swarm with a signed Mission Root intent.
    2. AIA Broker establishes an "Alignment Anchor" for the mission.
    3. Subagents periodically submit reasoning fragments to the AIA Broker for heartbeat attestation.
    4. AIA Broker verifies semantic alignment and issues a signed "Alignment Token."
    5. Tool calls require a valid, recent Alignment Token to proceed.

## 4. Design & Architecture

* **System Flow:**
    Subagent -> Reasoning Fragment -> AIA Broker -> [Semantic Similarity Engine] -> [Mission Manifest] -> Alignment Token -> Tool Gateway.

* **APIs / Interfaces:**
    * `POST /v1/align/heartbeat`: Submits a reasoning trace for alignment verification.
    * `GET /v1/align/status`: Retrieves the current alignment score for a session.
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) to persist Mission Root manifests and session alignment history.

## 5. Alternatives Considered

* **Binary Handshake Only:** Rejected because it doesn't account for drift *after* the initial delegation.
* **Full Context Monitoring:** Rejected due to high token overhead and latency; heartbeats provide a sampled, more efficient approach.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** All heartbeats must be hardware-attested (SRM/TPM) to prevent spoofing.
* **Observability:** Alignment scores are exported to the Unified Telemetry Bridge for real-time monitoring.

## 7. Evolutionary Changelog

* **2026-03-19:** Initial Document Creation.
