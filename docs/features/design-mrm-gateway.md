# Design Doc: Mission-Root Migration (MRM) Gateway
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Long-running AI agent swarms often need to shift their execution environment between local workstations, edge devices, and cloud sandboxes. However, current hardware-bound attestation models are host-specific, meaning that a "Mission Root" established on a local machine cannot be natively verified if the swarm migrates to the cloud. This leads to "Migration Stall" and loss of mission sovereignty.

The Mission-Root Migration (MRM) Gateway provides a secure protocol for migrating active mission roots and their hardware-attested state between hosts. It ensures that sovereignty is maintained across environmental shifts by creating a cryptographic "Lineage Handover" between host TPMs.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the secure transfer of mission roots between host environments.
    * Maintain hardware-attested lineage continuity during migration.
    * Provide a "Migration Attestation" token that proves valid handover.
    * Support seamless state resumption for migrated swarms.
* **Non-Goals:**
    * Migrating the raw model weights or runtime binaries (handled by container orchestration).
    * Bypassing local data residency laws.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile Agent Developer
* **Primary Goal:** Start a codebase review on a local laptop and seamlessly migrate it to a cloud-based GPU cluster for deep analysis without losing mission attestation.
* **The Happy Path (Tasks):**
    1. User starts the mission on a local host with a TPM-signed root.
    2. User triggers a migration to a cloud sandbox.
    3. The MRM Gateway performs a "Handover Handshake" between the local TPM and the cloud Secure Enclave.
    4. Mission state and RSL tokens are securely bundled and transferred.
    5. The cloud host resumes the mission, issuing a new attestation token that references the migrated local lineage.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        LocalHost[Local Host] -->|Initiate Migration| MRM[MRM Gateway]
        MRM -->|Handover Handshake| CloudHost[Cloud Host]
        LocalHost -->|Signed State Bundle| CloudHost
        CloudHost -->|Resume Mission| Swarm[Migrated Swarm]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mrm/export`: Bundle and sign mission state for export.
    * `POST /v1/mrm/import`: Verify and resume a migrated mission.
* **Data Storage/State:**
    * Migration manifests are stored in the Mesh-Resident Attestation Registry.

## 5. Alternatives Considered
* **Root Re-Attestation:** Rejected as it loses the cognitive lineage and requires manual user intervention at every shift.
* **Pure Cloud Sovereignty:** Rejected because it doesn't allow for local-first execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Migration requires multi-factor user attestation. State bundles are encrypted with the target host's public key.
* **Observability:** Migration events are logged in the "Mission-Root Migration Monitor."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
