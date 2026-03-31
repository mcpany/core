# Design Doc: Atomic Enclave Handover (AEH) Provider
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
With the introduction of Ephemeral Mission Enclaves (EME) in Claude Code v3.4.0, a new race condition (CVE-2026-44013) has emerged during teammate rotation. If a subagent terminates but the enclave identity is not atomically revoked, a successor can "Squat" on the remaining capabilities.

The AEH Provider ensures that the transfer of mission-root identities between specialized agents is an atomic, hardware-attested event.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent "Enclave Squatting" during teammate rotation.
    * Cryptographically link identity handovers to predecessor termination signals.
    * Support "Zero-Persistence" handovers where no state residue remains in the enclave.
* **Non-Goals:**
    * Managing the internal reasoning state of the enclave (handled by OpenClaw ContextEngine).
    * Providing long-term identity storage (identities are ephemeral by design).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Specialist Teammate
* **Primary Goal:** Securely hand over mission-root authority to a successor specialist without leaving "Capability Residue."
* **The Happy Path (Tasks):**
    1. Predecessor agent signals task completion to the AEH Provider.
    2. AEH Provider generates a hardware-attested **Handover Token**.
    3. Successor agent requests identity resumption using the token and its own TPM-bound identity.
    4. AEH Provider atomically wipes the predecessor's ephemeral keys and mints new ones for the successor.
    5. The mission enclave is "Re-Anchored" to the successor, and the predecessor's access is physically severed.

## 4. Design & Architecture
* **System Flow:**
    `[Predecessor] -> [Termination Signal] -> [AEH Hub] -> [Successor Handover] -> [Identity Re-Anchoring]`
* **APIs / Interfaces:**
    * `POST /v1/handover/initiate`: Issued by predecessor.
    * `POST /v1/handover/resume`: Issued by successor.
    * `X-AEH-Lineage-Link`: Header binding the handover to the specific mission-root UUID.
* **Data Storage/State:**
    * Volatile state stored in Secure Enclave (TPM NVRAM) for maximum isolation.

## 5. Alternatives Considered
* **Sequential Revocation (Rejected):** Creates a "Gap" window where neither agent owns the enclave, leading to cognitive stall.
* **Parallel Identities (Rejected):** Violates mission-root sovereignty by allowing two agents to hold the same authority simultaneously.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every handover event requires dual-agent hardware attestation.
* **Observability:** Handover latency and enclave wipe-success metrics are exported to the CSM.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
