# Design Doc: Attestation Batching Hub (ABH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The rapid shift toward multi-agent meshes has introduced a critical performance bottleneck: "Attestation Fatigue." High-trust swarms currently perform hardware-bound attestation (TPM/Secure Enclave) for every inter-agent tool call and state mutation. In high-density Agent Teams (e.g., Claude Code), this adds 200ms+ of coordination overhead per sub-task, driving subagents toward dangerous, un-attested speculative reasoning to maintain perceived performance.

The Attestation Batching Hub (ABH) evolves the MCP Any security layer to support aggregated cryptographic proofs. It enables the grouping of multiple task-level attestations into a single, hardware-signed mission-root token, drastically reducing the "Attestation Tax" without compromising the Zero-Trust integrity of the swarm.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce inter-agent coordination latency by 70% in high-frequency swarms.
    *   Provide hardware-attested "Batch Tokens" that represent a sequence of verified reasoning fragments.
    *   Neutralize "Latency Hijacking" by making secure attestation as fast as speculative reasoning.
    *   Support framework-neutral signature translation (Gemini to OpenClaw).
*   **Non-Goals:**
    *   Eliminating point-in-time attestation for critical mission-root initialization.
    *   Replacing the ARI Hub (ABH provides the transport-level signature, ARI Hub provides the semantic validation).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Security Architect
*   **Primary Goal:** Reduce the attestation overhead of a 10-teammate parallel mesh from 2.5s per task cycle to <500ms.
*   **The Happy Path (Tasks):**
    1.  The Mission Root agent initializes a session with the ABH, defining an "Attestation Window."
    2.  Parallel Teammates execute sub-tasks and submit "Soft-Attestation" signals to the ABH buffer.
    3.  ABH performs real-time semantic deconstruction and verification of the fragment lineage.
    4.  At the end of the window (or task threshold), ABH batches the verified fragments and issues a single TPM-signed "Batch Proof."
    5.  Teammates use the Batch Proof to commit state to the Blackboard in a single atomic operation, bypassing per-fragment hardware delays.

## 4. Design & Architecture
*   **System Flow:**
    [Teammate A] --(Fragment A)--> [ABH Proof Buffer] <--(Fragment B)-- [Teammate B]
                                        |
                                [Lineage Verification]
                                        |
                                [Hardware TPM Batching]
                                        |
                                        v
                            [TPM-Signed Mission Token]
                                        |
                                v               v
                          [Blackboard]      [Audit Log]

*   **APIs / Interfaces:**
    *   `POST /v1/abh/submit`: Submit reasoning fragments for batching.
    *   `GET /v1/abh/proof/{session_id}`: Retrieve the latest hardware-attested batch proof.
    *   `WS /v1/abh/stream`: High-speed stream for speculative fragment ingestion.
*   **Data Storage/State:**
    *   Secure memory-mapped buffers for pending fragments.
    *   Monotonic counter integration to prevent replay of batch proofs.

## 5. Alternatives Considered
*   **Trust Leases (LFTA):** While LFTA reduces per-call signatures, it provides a "Time-Bound" window of trust. ABH provides "Content-Bound" trust, which is more resilient to subagent hijacking within a lease window.
*   **Asynchronous Background Attestation:** Rejected as it allows "Poisoned Fragments" to be ingested by the agent before the hardware proof is ready, leading to hallucination cascades.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** If any fragment in a batch fails semantic validation (via ISD Hub), the entire batch proof is withheld, and a "Mesh Quarantine" signal is broadcast.
*   **Observability:** Visualized in the "Fast-Path Attestation Visualizer," showing the ratio of batched vs. individual signatures and real-time latency savings.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
