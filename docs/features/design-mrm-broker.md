# Design Doc: Mission-Root Migration (MRM) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agentic workflows move from local workstations to high-trust cloud enclaves (e.g., Claude Code v3.3.0), the "Mission Root" must be able to migrate across physical mesh boundaries without losing its hardware-attested state or cognitive continuity. Currently, migrating an active mission results in significant "Migration Jitter," where agents must re-read several megabytes of context, leading to 10s+ coordination stalls.

MCP Any needs to solve this by acting as the authoritative broker for cross-mesh handoffs. The MRM Broker will manage the cryptographic transfer of the mission identity and speculatively "pre-warm" target meshes with active context shards before the final handoff.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-locked transfer of Mission-Root identities between disparate device meshes.
    * Implement "State-Pre-warming" to reduce cognitive re-alignment jitter.
    * Maintain non-repudiable audit trails of the migration event.
    * Ensure Zero-Trust validation of target mesh security posture before handoff.
* **Non-Goals:**
    * Real-time mirroring of state across meshes (migration is a point-in-time handoff).
    * Supporting migration to non-hardware-attested target environments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile Agent Developer
* **Primary Goal:** Migrate a high-stakes coding mission from a local laptop to a secure cloud enclave for final deployment.
* **The Happy Path (Tasks):**
    1. User initiates migration via MCP Any CLI or UI.
    2. MRM Broker performs a security handshake with the target mesh, verifying TPM/SEP attestation.
    3. Target mesh is "pre-warmed" with the current mission-root intent and active context shards.
    4. MRM Broker cryptographically signs the "Migration Manifest" and transfers the hardware-bound root token.
    5. Mission resumes on the target mesh with <500ms cognitive latency.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Source MRM Broker->>Target MRM Broker: Migration Proposal (Hardware-Signed)
        Target MRM Broker->>Source MRM Broker: Attestation Receipt (TPM-Signed)
        Source MRM Broker->>Target MRM Broker: Context Shard Pre-warming (Encrypted Stream)
        Source MRM Broker->>Source Mission: Pause & Snapshot
        Source MRM Broker->>Target MRM Broker: Root Token Transfer (Encrypted)
        Target MRM Broker->>Target Mission: Resume & Re-align
    ```
* **APIs / Interfaces:**
    * `/v1/mrm/propose`: Initiates a migration request.
    * `/v1/mrm/attest`: Handles the cross-mesh hardware attestation exchange.
    * `/v1/mrm/prewarm`: Stream endpoint for context shard transfer.
* **Data Storage/State:**
    * Temporary "Migration Buffers" in shared memory (ZCMB-compatible).
    * Persistent migration logs in the local SQLite audit store.

## 5. Alternatives Considered
* **Manual Snapshot/Restore:** Rejected due to the loss of hardware attestation chain-of-custody.
* **Cloud-only Synchronization:** Rejected as it requires a central authority, violating the "Local Sovereignty" pillar.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Migration only allowed between meshes sharing a verified "Trust Root." All transfers are mTLS encrypted and hardware-signed.
* **Observability:** Migration events are logged as P0 security events in the Activity Map.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
