# Design Doc: Privacy-Preserving Audit (PPA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents become more autonomous, the need for auditability increases. However, traditional auditing involves exfiltrating raw reasoning traces and mission context, which often contains sensitive PII or trade secrets. This creates a "Security vs. Privacy" deadlock.

The Privacy-Preserving Audit (PPA) Hub is required to facilitate hardware-attested auditing of agent reasoning integrity without the exfiltration of sensitive mission context, utilizing Zero-Knowledge proofs.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the generation of Zero-Knowledge proofs for reasoning integrity (Gemini CLI PPRP pattern).
    * Provide a hardware-attested "Audit Seal" verifying that an agent followed mission-root constraints.
    * Allow third-party auditors to verify compliance without accessing raw context fragments.
* **Non-Goals:**
    * Storing raw reasoning traces for external access.
    * Replacing real-time semantic monitoring (AID Hub); it provides post-hoc or periodic attestation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Compliance Auditor
* **Primary Goal:** Verify that 1,000 automated agent sessions followed the "No Data Exfiltration" policy without reading the internal monologues of the agents.
* **The Happy Path (Tasks):**
    1. Agent sessions are completed, with each reasoning step generating a local "Integrity Hash".
    2. PPA Hub aggregates these hashes and generates a hardware-attested ZK-Proof against the Mission Policy.
    3. The Auditor requests the "Compliance Proof" for the mission.
    4. PPA Hub returns the ZK-Proof and the TPM signature of the Mission Root.
    5. Auditor's tools verify the proof against the public Policy manifest.
    6. Compliance is confirmed without a single byte of sensitive context leaving the local environment.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning] --> B[Integrity Hash Store]
        B --> C[PPA Hub]
        C --> D[ZK-Proof Generator]
        D --> E[Hardware-Signed Audit Seal]
        E --> F[Auditor]
        F -.->|Policy Verify| G[Public Policy Manifest]
    ```
* **APIs / Interfaces:**
    * `ppa.GenerateProof(missionID) -> ZKProof`: Generates a compliance proof for a specific mission.
    * `ppa.VerifySeal(auditSeal) -> Boolean`: Validates the TPM signature of an audit seal.
* **Data Storage/State:**
    * **Local Hash Vault:** Secure, volatile storage for reasoning integrity hashes.

## 5. Alternatives Considered
* **Centralized Logging:** Rejected due to the high risk of PII leakage and massive storage overhead.
* **Differential Privacy:** Useful for aggregate statistics, but insufficient for verifying individual agent compliance with strict security policies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Proofs are hardware-attested, ensuring they haven't been tampered with by a compromised agent.
* **Observability:** Integrates with the "Compliance Dashboard" for visualizing swarm-wide attestation status.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
