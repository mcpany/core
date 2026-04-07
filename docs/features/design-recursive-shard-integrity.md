# Design Doc: Recursive Shard Integrity (RSI) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from hierarchical to deep, horizontal meshes (e.g., OpenClaw v3.7.0), state management is becoming increasingly sharded and distributed. Current "Point-to-Point" attestation models fail in deep swarms where a state fragment may pass through 5+ specialist agents. If a mid-mesh agent is compromised, it can subtly "poison" the shard before passing it downstream. MCP Any needs a way to ensure the cryptographic lineage and integrity of these shards across the entire mission lifecycle.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested, hash-chained lineage for all context shards.
    * Enable sub-millisecond validation of shard integrity at any point in the mesh.
    * Support recursive "Parent-Child" attestation for multi-hop delegations.
* **Non-Goals:**
    * Encrypting the shard content itself (handled by the T2T Encryption Bridge).
    * Managing the storage of the shards (handled by the Universal Episodic Graph).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator
* **Primary Goal:** Verify that a context shard received from a 4th-hop specialist agent is a direct, untampered descendant of the mission-root intent.
* **The Happy Path (Tasks):**
    1. Primary Agent (Mission Root) initiates a mission and issues a root shard with a hardware-attested signature.
    2. Specialist A receives the shard, performs its task, and appends its reasoning fragment.
    3. Specialist A generates a new shard signature that hash-chains the previous root signature.
    4. Specialist B receives the chained shard and repeats the process.
    5. At the 4th hop, the RSI Hub validates the complete hash-chain against the hardware-attested mission-root public key.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Root Shard + Sig| SA[Specialist A]
        SA -->|Hash-Chained Sig| SB[Specialist B]
        SB -->|Hash-Chained Sig| SC[Specialist C]
        SC -->|Validation Request| RSI[RSI Hub]
        RSI -->|Hardware Attestation Check| TPM[TPM/SEP Enclave]
        TPM -->|Verified Lineage| RSI
        RSI -->|Success/Failure| SC
    ```
* **APIs / Interfaces:**
    * `POST /v1/rsi/sign`: Request a hash-chained signature for a new shard fragment.
    * `POST /v1/rsi/verify`: Verify the complete lineage of a sharded context fragment.
* **Data Storage/State:**
    * Shard metadata (signatures and lineage hashes) are stored in the hardware-attested Mesh-Resident Attestation Registry.

## 5. Alternatives Considered
* **Flat Multi-Signature:** Rejected due to O(N) token bloat in deep swarms.
* **Centralized State Hub:** Rejected to avoid single-point-of-failure and maintain mesh-resident performance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage is hardware-bound; even if a subagent's memory is compromised, it cannot forge a valid parent signature.
* **Observability:** RSI Hub logs every signature and verification event to the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
