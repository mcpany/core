# Design Doc: Hierarchical Intent Lease (HIL) v2
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale, the overhead of individual hardware-bound (TPM/SEP) attestation for every sub-task is becoming a performance bottleneck. Competitors like Claude Code have introduced "Lease Pooling" to mitigate this, but at the cost of significantly increasing the blast radius of a compromised sub-mission. If one agent in the pool is hijacked, it gains the full authority of the entire pool.

HIL v2 addresses this by implementing a hardware-locked, parent-child lease inheritance model. It allows for the efficiency of "pooled" attestation while maintaining the principle of least privilege. Sub-leases are mathematically restricted to a verified subset of the parent mission-root manifest, ensuring that a "pooled" capability cannot be abused by a subagent for which it was not intended.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable "Lease Pooling" to reduce hardware attestation latency by 60%+.
    * Enforce strict parent-child inheritance where sub-leases are subsets of the mission-root manifest.
    * Implement "Atomic Lease Reclamation" for sub-millisecond revocation of specific sub-leases.
    * Provide cryptographic proof of sub-lease lineage back to the hardware-attested mission root.
* **Non-Goals:**
    * Managing the primary model reasoning loop.
    * Providing transport-layer encryption (handled by the T2T Encryption Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Group 5 specialized sub-tasks (Linter, Security Scan, Unit Test, Docs, Deploy) under a single pooled lease while ensuring the "Docs Specialist" cannot use the "Deploy Specialist" hardware-bound secrets.
* **The Happy Path (Tasks):**
    1. The mission-root initiates a `CreatePooledLease` request with the combined manifest for the 5 sub-tasks.
    2. MCP Any performs a hardware handshake (TPM) to sign the root pool.
    3. As the "Docs Specialist" agent is spawned, the orchestrator requests a `DeriveSubLease` for only the "FileSystem:Read:Docs" capability.
    4. HIL v2 validates that the requested capability is a subset of the pooled lease and issues a time-bound, hardware-locked sub-token.
    5. The Docs specialist attempts to access "Deploy:Secrets"; the HIL v2 provider interdicts the call as it exceeds the sub-lease scope, even though the secret is available in the broader pool.
    6. Upon task completion, the sub-lease is atomically reclaimed.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [HIL v2 Provider] -> [Hardware Enclave (TPM)] -> [Lease Pool] -> [Sub-Lease derivation] -> [Subagent]`
* **APIs / Interfaces:**
    * `hil.v2.CreatePooledLease(manifest MissionManifest) (PoolToken, error)`
    * `hil.v2.DeriveSubLease(parentPool PoolToken, scope CapabilitySubset) (SubLeaseToken, error)`
    * `hil.v2.RevokeSubLease(subLease SubLeaseToken) error`
* **Data Storage/State:**
    * Lease states are managed in the RAMS-compliant Blackboard with hardware-bound versioning.
    * Active pool/sub-lease mappings are held in secure, kernel-resident memory.

## 5. Alternatives Considered
* **Flat Lease Pooling:** Rejected due to high security risk (unrestricted authority within the pool).
* **Per-Call Attestation:** Rejected due to "Attestation Tax" (100ms+ latency per sub-task).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HIL v2 uses monotonic counters and hardware-bound lineage tokens to prevent "Attestation Replay" and "Logic Grafting" attacks.
* **Observability:** All derivation and reclamation events are logged with high-resolution timestamps in the "Mission Lease Manager" dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
