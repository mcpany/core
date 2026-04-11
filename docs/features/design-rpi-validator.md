# Design Doc: Reasoning-Path Integrity (RPI) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the discovery of **Reasoning Grafting** (CVE-2026-88102), the "Universal Agent Bus" faces a critical threat where malicious subagents can append or inject unauthorized instructions into hardware-signed reasoning fragments during context summarization or handoff. Existing fragment-level validation (ARI) ensures individual segments are valid but fails to protect the **continuity** and **lineage** of the reasoning path. The **RPI Validator** implements real-time semantic hash-chaining to ensure that every fragment in a coordination stream is cryptographically and semantically linked to its predecessor and the hardware-attested mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Semantic Hash-Chaining" for inter-agent reasoning fragments.
    * Neutralize "Reasoning Grafting" attacks by verifying path continuity.
    * Provide hardware-bound attestation for the entire "Chain of Thought" lineage.
    * Support sub-millisecond validation latency to prevent cognitive stall.
* **Non-Goals:**
    * Does not perform the actual summarization (delegated to ContextEngine).
    * Does not replace individual fragment validation (ARI).
    * Will not store the full reasoning history (delegated to the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure that a "summarized" reasoning trace handed off between agents has not been tampered with or polluted by intermediate specialist subagents.
* **The Happy Path (Tasks):**
    1. The primary agent generates an initial reasoning fragment and signs it with the mission-root key.
    2. As subagents contribute fragments, the RPI Validator generates a semantic hash of the new fragment, combined with the previous fragment's hash.
    3. Each link in the chain is hardware-attested via the TPM.
    4. During summarization, the RPI Validator verifies the entire hash-chain against the mission-root manifest.
    5. If any "grafted" or unauthorized instructions are detected (broken chain), the handoff is interdicted and the session is quarantined.

## 4. Design & Architecture
* **System Flow:**
    - `Agent` -> `RPI Validator` (Submit Fragment + Previous Hash)
    - `RPI Validator` -> `Semantic Engine` (Extract semantic intent fingerprint)
    - `RPI Validator` -> `TPM` (Generate signed hash-chain link)
    - `RPI Validator` -> `Blackboard` (Store link metadata)
    - `ContextEngine` -> `RPI Validator` (Validate chain before summarization)
* **APIs / Interfaces:**
    - `POST /rpi/v1/chain/link`: Append a validated fragment to the reasoning chain.
    - `GET /rpi/v1/chain/verify`: Validate the integrity of a summarized path.
* **Data Storage/State:**
    - Uses an append-only "Lineage Ledger" stored in the hardware-protected region of the Blackboard.

## 5. Alternatives Considered
* **Full-Trace Signing:** Rejected due to O(N^2) token bloat and performance degradation in deep swarms.
* **Periodic Checkpointing:** Rejected as it leaves "Grafting Windows" between checkpoints.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    - Semantic hashes are salt-protected and bound to the hardware Environment ID (EAP-compliant).
    - Prevents "Trace Replay" by including monotonic session counters in each hash link.
* **Observability:**
    - Export "Lineage Integrity" metrics to the System Health Dashboard.
    - Alert on "Chain Rupture" events with high-fidelity semantic diffs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
