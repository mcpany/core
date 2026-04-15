# Design Doc: Sovereign Shard Migration (SSM) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from local-only environments to heterogeneous meshes (Local + Edge + Cloud), the need to migrate cognitive state without losing integrity or security has become critical. Today's "Sovereign Shard Migration" (SSM) patterns in OpenClaw reveal that "Physical Shard Sovereignty" must be mobile.

The SSM Hub in MCP Any will act as the authoritative broker for migrating hardware-attested context shards between disparate nodes. It ensures that an agent's "memory" can follow its "compute" (e.g., moving a reasoning task from a desktop to a secure cloud enclave) while maintaining a continuous, untamperable cryptographic lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested migration of "Entangled State Shards" between heterogeneous nodes.
    * Maintain cryptographic continuity of mission-root intents during migration.
    * Neutralize "Migration Shadowing" where a rogue node attempts to intercept or spoof a shard during transit.
    * Support seamless state resumption on the target node with sub-100ms latency.
* **Non-Goals:**
    * Providing long-term cold storage for shards (handled by Skill-State Sovereignty Broker).
    * Migrating underlying model weights or agent binaries.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Migrate a high-trust database-audit shard from a local workstation to a production cloud enclave for final verification.
* **The Happy Path (Tasks):**
    1. The supervisor agent initiates a migration request via the SSM Hub, identifying the shard and the target node ID.
    2. The SSM Hub verifies the mission-root authority and generates a migration-locked "Handover Token."
    3. The source node cryptographically "seals" the shard against the target node's hardware identity (TPM/SEP).
    4. The shard is transmitted over an Attested Mesh Tunnel (AMT).
    5. The target node verifies the Handover Token and Handshake Lineage before unsealing the shard into its secure memory region.
    6. The SSM Hub records the successful migration in the Mesh-Resident Lineage Tracker.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant S as Source Node (Local)
        participant H as SSM Hub (Orchestrator)
        participant T as Target Node (Cloud)

        S->>H: Migration Intent (Shard_ID, Target_ID, Mission_Root)
        H->>H: Validate Intent & Quorum
        H->>T: Pre-flight Challenge (Target_Identity_Proof)
        T-->>H: Identity Attestation (TPM-Signed)
        H->>S: Grant Migration (Migration_Key, Target_PubKey)
        S->>S: Seal Shard for Target
        S->>T: Transmit Shard (via AMT)
        T->>H: Migration Complete Notification
        H->>H: Update Lineage & Mesh State
    ```
* **APIs / Interfaces:**
    * `POST /v1/ssm/initiate`: Initiate a migration branch.
    * `POST /v1/ssm/verify`: Verify target node readiness and hardware identity.
    * `GET /v1/ssm/status/{migration_id}`: Track migration lifecycle.
* **Data Storage/State:**
    * Migration metadata is stored in the Mesh-Resident Attestation Registry.
    * Shards remain in-memory (DME-locked) during the migration process.

## 5. Alternatives Considered
* **Stateless Re-computation**: Forcing the target node to re-reason based on raw logs. Rejected due to prohibitive token costs and "Hallucination Drift" between model versions.
* **Simple Encrypted Transport**: Using standard TLS without hardware attestation. Rejected because it cannot prove the target node hasn't been compromised at the OS level (Neutralized by TPM-signed proof).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All migrations require hardware-bound identity verification. Shards are never decrypted outside of a verified secure enclave.
* **Observability:** Migrations are logged with monotonic counters in the Command Traceability Provider (CTP).

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
