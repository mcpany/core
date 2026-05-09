# Design Doc: NHI Lifecycle Governance Provider
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As agent swarms evolve from ephemeral task-based sessions into persistent autonomous service meshes, the risk profile of Non-Human Identities (NHI) shifts from simple access control to complex lifecycle management. Static, hardware-bound tokens are vulnerable to "Session Squatting" if an agent is compromised or if a mission outlives its intended duration.

The NHI Lifecycle Governance Provider acts as the central authority for managing the "Birth-to-Death" cycle of agent identities within the Universal Agent Bus. It ensures that every agent is bound to a mission-specific identity that rotates automatically and is forcefully revoked upon task completion or mission termination.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement automated, mission-bound identity rotation for all connected agents.
    * Provide a hardware-locked revocation mechanism (integrating with TPM/Secure Enclave) for NHIs.
    * Enforce "Continuous Attestation" where agents must periodically prove their mission alignment.
    * Support "Post-Mission Sanitization" of identity metadata to prevent lineage leaks.
* **Non-Goals:**
    * Replacing low-level transport security (e.g., mTLS).
    * Managing human identities (handled by existing IAM providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Compliance Officer
* **Primary Goal:** Ensure that a specialized "Database-Admin" agent cannot perform unauthorized actions after its assigned maintenance task is complete.
* **The Happy Path (Tasks):**
    1. The NHI Lifecycle Provider issues a task-bound identity token to the agent at mission start.
    2. The token is cryptographically linked to the specific "Maintenance Mission" manifest.
    3. The agent performs its authorized tool calls, with the identity rotating every 30 minutes via hardware-bound handshakes.
    4. Upon completion of the maintenance task, the Mission Root signals "Task Termination."
    5. The NHI Lifecycle Provider immediately broadcasts a hardware-locked revocation signal.
    6. All subsequent tool calls from that agent identity are rejected by the gateway.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Start] --> B[NHI Minting]
        B --> C[Hardware Binding]
        C --> D[Active Agent Session]
        D --> E{Interval Met?}
        E -- Yes --> F[Identity Rotation]
        F --> D
        E -- No --> G{Mission End?}
        G -- Yes --> H[Hardware-Locked Revocation]
        H --> I[Identity Metadata Purge]
    ```
* **APIs / Interfaces:**
    * `MintMissionIdentity(manifest, rootToken) -> NHIToken`
    * `RotateIdentity(currentToken) -> NewNHIToken`
    * `RevokeIdentity(nhitokenID) error`
* **Data Storage/State:**
    * Identity state is stored in a TEE-protected segment of the Blackboard, ensuring that even a compromised host cannot inspect the rotation keys.

## 5. Alternatives Considered
* **Short-Lived JWTs**: Rejected as they lack hardware-bound non-repudiability and cannot be proactively revoked if the token is stolen before expiry.
* **Manual Revocation**: Rejected due to the "Machine-Speed" requirement of autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Integrates with the ARL Provider for mesh-wide revocation broadcasting.
* **Observability:** Identity rotation and revocation events are surfaced in the "Mesh Identity Manager" UI.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
