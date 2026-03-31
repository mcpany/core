# Design Doc: Migration-Aware State Sanitizer (MASS)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the introduction of Dynamic Mesh Resilience (DMR) in OpenClaw, agent swarms can now re-shard and migrate "Entangled State" between physical nodes to handle failures. However, this has introduced "Migration Smuggling," where malicious subagents hide unauthorized context fragments within the encrypted migration payloads. Current sanitization fails because fragments are often re-composed only after transit.

MCP Any needs to solve this by acting as the authoritative "Re-Attestation Gate" during migrations. MASS will ensure that no state fragment crosses node boundaries unless it is semantically verified against the mission-root intent, even when encapsulated in encrypted DMR tunnels.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all DMR state migration signals.
    * Perform real-time semantic sanitization of fragments *before* they are re-encrypted for transit.
    * Validate fragment lineage against the hardware-attested mission root.
    * Provide cryptographic proof of "Clean Migration" to the recipient node.
* **Non-Goals:**
    * Replacing the underlying transport encryption used by DMR.
    * Managing the physical node selection for re-sharding.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Prevent "Ghost Fragments" from being smuggled into a high-trust enclave during a fail-over event.
* **The Happy Path (Tasks):**
    1. A specialist subagent node detects hardware failure and initiates a DMR migration.
    2. MCP Any's MASS middleware intercepts the migration request.
    3. MASS decrypts the state fragment within a secure enclave (HAPE).
    4. The fragment is scanned for "Smuggled Intents" diverging from the root mission.
    5. Clean fragments are re-signed with a "Migration-Attestation" token.
    6. The state is successfully migrated to the new node and accepted by its local gateway.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Node A] --(Migration Request)--> [MASS Hub] --(Sanitize)--> [MASS Hub] --(Attested State)--> [Subagent Node B]
* **APIs / Interfaces:**
    * `rpc MigrateState(StateFragment) returns (AttestedFragment)`
    * `rpc VerifyMigrationToken(Token) returns (bool)`
* **Data Storage/State:**
    * Temporary secure memory buffers for fragment scanning.
    * Mission-root intent manifest for semantic comparison.

## 5. Alternatives Considered
* **Post-Migration Scanning:** Rejected due to "Time-of-Ingestion" risk; the new node reasoning engine could ingest the poisoned fragment before the scan completes.
* **Zero-Knowledge Migration Proofs:** Considered but rejected for initial version due to 500ms+ overhead; MASS semantic scanning is 10x faster.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MASS operates in a "Deny-by-Default" mode during migration; any fragment with ambiguous lineage is quarantined.
* **Observability:** Migration-sanitization events are logged to the "Mesh Resilience Dashboard" with drift scores.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
