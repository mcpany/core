# Design Doc: Mesh-Resident Attestation Registry (MRAR)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
The emergence of "Lineage Hijacking" and the persistence of unauthenticated inter-agent coordination prove that the transport layer alone is insufficient for securing agent swarms. As teammates from disparate frameworks coordinate via shared mailboxes and sharded meshes, the "Universal Agent Bus" needs a central, hardware-attested registry to manage identity fragments and their authorized environmental bounds.

## 2. Goals & Non-Goals
* **Goals:**
    * Act as the authoritative registry for all hardware-attested agent identities in the mesh.
    * Enforce environmental sovereignty by binding identities to specific process/container bounds.
    * Provide sub-millisecond identity verification for inter-teammate coordination.
    * Support "Mission-Root" anchoring for all identity fragments.
* **Non-Goals:**
    * Providing long-term identity storage across multiple independent missions (identities are session-bound).
    * Managing human identity (this is an NHI-only registry).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Verify the hardware-attested identity and environmental bounds of a specialized teammate before delegating a high-risk tool call.
* **The Happy Path (Tasks):**
    1. A specialized teammate spawns and registers its hardware-attested identity fragment with the MRAR.
    2. The MRAR validates the teammate's TPM signature and its environmental bounds (PID/Cgroup).
    3. The Mission-Root agent queries the MRAR to verify the teammate's lineage before delegation.
    4. The MRAR provides a "Sovereign Identity Token" cryptographically bound to the mission-root.
    5. The teammate uses the token to authenticate its coordination messages in the shared mailbox.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        T[Teammate Agent] -->|Register Identity| MRAR[Mesh-Resident Attestation Registry]
        MRAR -->|Verify TPM/Env| TPM[Hardware TPM/Enclave]
        MR[Mission Root] -->|Query Lineage| MRAR
        MRAR -->|Issue Sovereign Token| MR
        MR -->|Delegate Task| T
        T -->|Auth Coordination| Mailbox[Shared Mailbox]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mrar/identity/register`: Register a hardware-attested identity fragment.
    * `GET /v1/mrar/identity/verify`: Verify an identity fragment and its environmental bounds.
    * `GET /v1/mrar/lineage/trace`: Retrieve the hardware-attested lineage of a specific identity.
* **Data Storage/State:**
    * Identities are stored in an in-memory, hardware-encrypted registry indexed by mission-root session IDs.

## 5. Alternatives Considered
* **Flat Token-Based Identity:** Rejected as it cannot prevent token reuse across unauthorized environments.
* **Centralized Cloud IAM:** Rejected due to the latency requirements of local, high-frequency coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All registry entries are hardware-attested and mission-bound. Identity fragments expire automatically upon mission termination.
* **Observability:** Real-time visualization of the mesh identity graph via the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
