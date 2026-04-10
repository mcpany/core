# Design Doc: Autonomous Mission Resumption (AMRA) Hub
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms evolve toward "Long-Haul Agency," missions frequently exceed the 24-hour mark. Current security architectures rely on session-bound hardware-attested tokens which naturally decay or expire, leading to "Cognitive Stall" when agents lose the ability to verify their mission-root authority. MCP Any needs a standardized, hardware-locked mechanism to resume missions securely after token decay without requiring a full manual re-attestation.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-100ms mission resumption for long-running agent swarms.
    * Mandate hardware-bound (TPM) monotonic counters to prevent replay attacks during resumption.
    * Ensure mission-root sovereignty is maintained across session boundaries.
* **Non-Goals:**
    * Replacing the initial mission attestation process.
    * Providing long-term storage for agent reasoning monologues (handled by SRM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-day Swarm Orchestrator
* **Primary Goal:** Resume a 48-hour data-migration mission after a session-token expiry without manual intervention.
* **The Happy Path (Tasks):**
    1. The parent agent detects a "Session Token Expired" signal from the transport layer.
    2. The parent agent requests a "Resumption Challenge" from the AMRA Hub.
    3. The AMRA Hub issues a challenge tied to the hardware-bound monotonic counter of the mission-root.
    4. The parent agent signs the challenge using the Mission-Root's TPM-resident key.
    5. The AMRA Hub validates the signature and counter, issuing a new set of hardware-attested session tokens for the mesh.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `detects expiry` -> `AMRA Hub (Challenge Request)` -> `TPM (Sign Challenge + Monotonic Counter)` -> `AMRA Hub (Verify & Re-issue)` -> `Mesh (Resume Operations)`
* **APIs / Interfaces:**
    * `POST /v1/amra/challenge`: Request a resumption challenge.
    * `POST /v1/amra/verify`: Submit signed challenge and monotonic proof.
* **Data Storage/State:**
    * Uses the TPM monotonic counter as the authoritative state.
    * Ephemeral mission-root lineage map stored in the `Mesh Identity Registry`.

## 5. Alternatives Considered
* **Extended Token Lifetimes:** Rejected due to increased risk of "Stale-Token Hijacking."
* **Manual Re-attestation:** Rejected due to the "Cognitive Stall" it introduces in autonomous workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory TPM signatures prevent software-only spoofing of mission resumption.
* **Observability:** Every resumption event is logged in the `Sovereign Audit Trail` with the associated monotonic counter value.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.

### Update: 2026-07-25 - Supporting Asynchronous Teammate Handoffs (ATH)
**Context:** Today's research revealed Claude Code's "ATH" standard for multi-node missions. Agents now need to move mission state across disparate hardware nodes while maintaining hardware-attested continuity.
**Architecture Adjustment:**
* Section 4 now includes the "Teammate-Resume-Token" (TRT) as a core artifact of the AMRA Hub.
* The AMRA Hub will now act as a global (cross-node) broker for TRTs, allowing a session paused on Node A to be resumed on Node B using a hardware-locked attestation bridge.
**Security Impact:** Prevents "Cognitive Amnesia" and unauthorized state hijacking during multi-node handoffs by binding the TRT to a verifiable hardware lineage.
