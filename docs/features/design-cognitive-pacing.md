# Design Doc: Cognitive Pacing Middleware
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
As agent swarms become more parallel and hardware-accelerated, subagent reasoning frequency can sometimes exceed the speed at which the parent agent or the underlying hardware security module (TPM) can perform attestation. This leads to "Attestation Race Conditions" and "Cognitive Stall," where the mesh becomes unstable due to unsynchronized trust signals.

MCP Any needs to implement Cognitive Pacing to synchronize subagent reasoning with the mesh's attestation capacity.

## 2. Goals & Non-Goals
* **Goals:**
    * Throttle subagent reasoning frequency to match the hardware-attestation speed of the mesh.
    * Eliminate race conditions between reasoning loops and trust signatures.
    * Provide a standardized "Pacing Signal" for heterogeneous frameworks (OpenClaw, Claude Code).
    * Dynamically adjust pacing based on real-time TPM latency.
* **Non-Goals:**
    * Modifying the internal logic of the subagent's reasoning engine.
    * Providing a global rate-limiter for cost control (that is handled by the Quota Monitor).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Ensure 50 specialist subagents can coordinate without triggering attestation failures due to signal overlapping.
* **The Happy Path (Tasks):**
    1. Orchestrator enables "Cognitive Pacing" for the mission.
    2. Subagents begin parallel execution.
    3. Pacing Middleware monitors the hardware-attestation queue.
    4. Middleware sends "Wait-for-Attestation" signals to subagents when queue depth exceeds threshold.
    5. Subagents pause until the "Attestation Ready" signal is received.
    6. Swarm completes the task with 100% hardware-attested integrity and zero stalls.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Loop] -> [Pacing Hook] -> [Attestation Queue Monitor] -> [TPM/HSM]
    The middleware acts as a semaphore for the reasoning loop, tied to the hardware attestation lifecycle.
* **APIs / Interfaces:**
    * `pacing.CheckAttestationAvailability(sessionID) -> bool`: Checks if the TPM is ready for a new signature.
    * `pacing.SetPacingPolicy(sessionID, policy)`: Configures latency-bound vs. throughput-bound pacing.
* **Data Storage/State:**
    Real-time attestation latency metrics are stored in memory and exposed via the UI roadmap features.

## 5. Alternatives Considered
* **Asynchronous Attestation:** Allowing reasoning to proceed while signatures are generated in the background. Rejected because it violates Zero-Trust (actions would be taken before being fully verified).
* **Static Throttling:** Fixed delay between turns. Rejected as it is too slow for low-latency TPMs and too fast for high-load scenarios.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pacing ensures that no reasoning fragment is ingested without a valid, fresh hardware signature.
* **Observability:** Integrated with the "Cognitive Pacing Monitor" in the UI for real-time visualization of sync status.

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
