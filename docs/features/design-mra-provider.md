# Design Doc: Mesh-Resident Attestation (MRA) Provider
**Status:** Draft
**Created:** 2026-06-12

## 1. Context and Scope
As agent swarms move toward ARI-v2 (Active Reasoning Interdiction) for protecting mission integrity, the security of the underlying semantic hashes has become the primary attack surface. Malicious subagents are attempting to "spoof" reasoning fragments by exploiting hash-collision vulnerabilities in legacy middleware or by replaying previously authorized fragments.

The Mesh-Resident Attestation (MRA) Provider provides the hardware-bound infrastructure required for ARI-v2. It utilizes local Trusted Platform Modules (TPM) or Secure Enclaves to generate and verify collision-resistant semantic hashes, ensuring that every fragment in the "Reasoning Mainline" is cryptographically tied to the physical hardware that produced it.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-accelerated service for generating and verifying semantic hashes.
    * Mandate collision-resistant hashing (e.g., TPM-backed SHA-384) for all ARI fragments.
    * Implement "Freshness Attestation" to prevent reasoning replay attacks.
    * Act as the mesh-resident root of trust for coordination hashes.
* **Non-Goals:**
    * Performing semantic analysis of the reasoning itself (handled by the ARI Hub).
    * Managing inter-agent transport encryption (handled by T2T Bridge).
    * Long-term storage of reasoning fragments (handled by the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure that a "Logic Grafting" attempt using a spoofed hash is detected and blocked by the hardware layer.
* **The Happy Path (Tasks):**
    1. Parent Agent generates a reasoning fragment.
    2. ARI Hub requests a hardware-attested hash from the MRA Provider.
    3. MRA Provider utilizes the local TPM to sign the fragment's semantic content and issues a "Hardware Hash Token."
    4. Specialist Agent receives the fragment and token.
    5. Specialist Agent proposes a follow-up fragment, including the parent's "Hardware Hash Token."
    6. MRA Provider verifies the parent token's signature and freshness using the TPM root.
    7. If verification succeeds, a new hash-chain link is issued.
    8. If a collision or spoof is detected, the TPM rejects the request and triggers a mesh-wide quarantine.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[ARI Hub] --> B[MRA Provider]
        B --> C[TPM/Secure Enclave]
        C --> D[Hardware Hash Generator]
        D --> E[Attestation Receipt]
        E --> B
        B --> A
        F[Fragment Proposal] --> G[Hash Verifier]
        G --> C
        C -- Valid --> H[Commit to Lineage]
        C -- Invalid --> I[Trigger Quarantine]
    ```
* **APIs / Interfaces:**
    * `mra.AttestFragment(fragment, salt) -> AttestationToken`
    * `mra.VerifyToken(token, fragment) -> bool`
* **Data Storage/State:**
    * **Nonce Registry:** A transient, in-memory registry of unique nonces for freshness attestation.
    * **Hardware Key Store:** Securely managed within the local TPM.

## 5. Alternatives Considered
* **Software-Only SHA-256:** Rejected due to vulnerability to collision spoofing and lack of hardware-bound non-repudiation.
* **Remote Attestation (Cloud-based):** Rejected due to high latency (150ms+) which violates the "Machine-Speed" mesh requirement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The MRA Provider is the foundation of Zero-Trust coordination. It must be isolated from the LLM reasoning process.
* **Observability:** Integrated with the "TPM Security Monitor" for real-time hardware health and attestation logs.

## 7. Evolutionary Changelog
* **2026-06-12:** Initial Document Creation. Supporting ARI-v2 requirements for collision-resistant, hardware-bound reasoning hashes.
