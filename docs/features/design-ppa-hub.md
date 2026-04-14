# Design Doc: Privacy-Preserving Audit (PPA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent meshes become more distributed and autonomous, the need for external auditing (by humans or compliance agents) increases. However, exposing the raw context of a mission-root reasoning chain to an auditor creates a massive privacy and security risk. Today's market shift toward "Cross-Mesh Auditability" (e.g., Gemini CLI v0.59.0) proves that we need a way to attest to the integrity of an agent's reasoning without revealing the sensitive data within that reasoning.

The Privacy-Preserving Audit (PPA) Hub provides the infrastructure for generating and verifying hardware-attested Zero-Knowledge reasoning proofs across the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the generation of Zero-Knowledge proofs for reasoning integrity.
    * Allow external auditors to verify that a reasoning chain followed mission-root constraints without seeing raw context fragments.
    * Provide hardware-attested (TPM) signatures for all generated proofs.
    * Support "Cross-Mesh Auditability" where reasoning traces span multiple distributed nodes.
* **Non-Goals:**
    * Providing a full text-based audit log (this is what PPA aims to avoid for sensitive data).
    * Acting as a primary reasoning engine; it is a verification layer.

## 3. Critical User Journey (CUJ)
* **User Persona:** Compliance Officer / Auditor
* **Primary Goal:** Verify that a multi-agent swarm task involving sensitive financial data adhered to the "No External Exfiltration" policy without viewing the actual financial records.
* **The Happy Path (Tasks):**
    1. The mission-root agent completes a high-stakes task across multiple nodes.
    2. The PPA Hub collects hash-chained reasoning fragments from all involved nodes (AMT tunnels).
    3. The PPA Hub generates a Zero-Knowledge Proof (ZKP) that the reasoning steps remained within the policy-bound sandbox.
    4. The auditor requests the proof via the "Zero-Knowledge Audit Viewer" in the UI.
    5. The PPA Hub provides the TPM-signed ZKP and a cryptographic receipt.
    6. The auditor's verification tool confirms the proof is valid against the mission manifest.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Reasoning] --> B[ARI Hub (Fragment Hashing)]
        B --> C[PPA Hub]
        D[Mission Manifest] --> C
        C --> E[ZKP Generator]
        E --> F[TPM Signing]
        F --> G[Auditor Verification]
    ```
* **APIs / Interfaces:**
    * `ppa.GenerateReasoningProof(missionID, policyID) -> ProofID`: Initiates ZKP generation.
    * `ppa.VerifyProof(proofID, hardwareSignature) -> Boolean`: Validates a proof.
    * `ppa.GetAuditManifest(missionID) -> Manifest`: Returns the list of policies being audited.
* **Data Storage/State:**
    * **Proof Registry:** SQLite database storing proof metadata and hardware signatures.
    * **Policy Store:** Repository of verifiable mission constraints.

## 5. Alternatives Considered
* **Differential Privacy (DP):** Rejected because DP introduces noise that can degrade the auditability of precise agent actions. ZKPs provide deterministic verification.
* **Encrypted Log Storage:** Rejected because it still requires sharing decryption keys with the auditor, which is a "all or nothing" approach to privacy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The PPA Hub itself must run in a hardware-isolated enclave to ensure that the raw context it uses to generate proofs is never exposed to the host OS.
* **Observability:** Integrated with the "Zero-Knowledge Audit Viewer" and "Mission-Root Lineage Tracker" for visual verification.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
