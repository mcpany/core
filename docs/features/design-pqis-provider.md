# Design Doc: Post-Quantum Intent Signatures (PQIS) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent missions become increasingly long-lived and high-value, the "Mission Root" intent becomes a target for long-term data harvesting. Traditional ECDSA signatures, while secure today, are vulnerable to future quantum computing attacks (Store Now, Decrypt Later). MCP Any must ensure that the cryptographic anchors of agent sovereignty are future-proofed against quantum-enabled decryption and forgery.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement NIST-standard Post-Quantum Cryptography (PQC) for mission-root signing.
    * Support hybrid signatures (RSA/ECDSA + ML-KEM) to maintain backward compatibility.
    * Provide hardware-attested key generation for PQC primitives using supported TPMs.
* **Non-Goals:**
    * Migrating all legacy MCP tool signatures (focused on the Mission Root).
    * Providing general-purpose PQC encryption for tool data.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Ensure that a sensitive 6-month mission intent remains sovereign and untamperable for its entire duration, even in a post-quantum landscape.
* **The Happy Path (Tasks):**
    1. Architect initializes a new Mission Root with a "Quantum-Resistant" policy.
    2. MCP Any generates a hybrid (ECDSA + SLH-DSA) keypair in the local Secure Enclave.
    3. The primary intent is signed with both keys and broadcast to the teammate mesh.
    4. Specialist agents verify the hybrid signature; those with PQ-capabilities verify the SLH-DSA fragment.
    5. The mission continues with the guarantee that the intent anchor is quantum-secure.

## 4. Design & Architecture
* **System Flow:**
    [User Intent] -> [PQIS Provider] -> [TPM/Secure Enclave (ML-KEM/SLH-DSA)] -> [Hybrid Signature Token] -> [Universal Agent Bus]
* **APIs / Interfaces:**
    * `SignIntent(intent_id, payload, policy_level)` -> `pq_signature_token`
    * `VerifyIntent(intent_id, token)` -> `boolean`
* **Data Storage/State:**
    * PQ public keys are stored in the Shared KV Store (Blackboard) under the `system.pqis` namespace.

## 5. Alternatives Considered
* **Waiting for Hardware Support:** Rejected because software-based PQC (liboqs) is sufficiently performant for the low-frequency signing of mission roots.
* **Pure PQC:** Rejected in favor of hybrid signatures to ensure current-generation agents can still verify the basic lineage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PQC keys must be hardware-bound to prevent exfiltration.
* **Observability:** Log the algorithm version used for every signature to facilitate future migration.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
