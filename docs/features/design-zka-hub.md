# Design Doc: Zero-Knowledge Audit (ZKA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Enterprise adoption of autonomous agent swarms is often blocked by "Audit Exhaustion"—the need for human review of thousands of reasoning traces—or "Privacy Paradox"—the need to audit agents without exposing the sensitive mission context (e.g., PII or proprietary code) contained in those traces.

The Zero-Knowledge Audit (ZKA) Hub is required to facilitate hardware-attested proofs of reasoning integrity, allowing third-party auditors or parent supervisors to verify that an agent followed safety protocols without exfiltrating raw context fragments.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate Zero-Knowledge (ZK) proofs for agent reasoning paths and tool usage.
    * Enable "Privacy-Preserving Auditing" where auditors only see the "Safety Pass/Fail" and proof of integrity.
    * Anchor ZK proofs to hardware enclaves (TPM/SEP) to prevent proof spoofing.
    * Support "Streaming Audit" where proofs are generated incrementally during long-running missions.
* **Non-Goals:**
    * Redacting data in the reasoning engine (handled by RAR).
    * Providing a general-purpose ZK-prover for non-agentic tasks.
    * Storing raw reasoning data; it only stores the proofs and their associated metadata.

## 3. Critical User Journey (CUJ)
* **User Persona:** Compliance Officer
* **Primary Goal:** Verify that a "Health Data Researcher" agent did not exfiltrate patient names while querying a medical database, without seeing the patient data itself.
* **The Happy Path (Tasks):**
    1. Agent performs the research task, utilizing the RAR engine to mask PII.
    2. ZKA Hub monitors the reasoning trace and the tool outputs.
    3. The Hub generates a ZK-proof asserting: "The agent called Tool X with parameters matching Policy Y, and no string matching Pattern Z was found in the output."
    4. The proof is signed by the hardware TPM.
    5. The Compliance Officer reviews the proof in the "ZKA Audit Viewer."
    6. The Hub provides a "Cryptographic Receipt" of compliance without the Officer ever seeing a patient name.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Trace] --> B[ZKA Prover]
        C[Security Policy] --> B
        B --> D[ZK-Proof Generation]
        D --> E[TPM Signing]
        E --> F[ZKA Hub Storage]
        F --> G[Auditor UI]
    ```
* **APIs / Interfaces:**
    * `zka.GenerateProof(traceID, policyID) -> ProofToken`: Initiates proof generation.
    * `zka.VerifyProof(proofToken) -> bool`: Validates a proof against the public policy.
    * `zka.ExportAuditReceipt(proofToken) -> Receipt`: Generates a verifiable compliance document.
* **Data Storage/State:**
    * **Proof Registry:** Ledger of ZK-proofs and their hardware-attestation status.
    * **Policy Store:** Repository of audit-able security rules defined in Rego or CEL.

## 5. Alternatives Considered
* **Full Logging with Manual Review:** Rejected due to "Privacy Paradox" and high operational cost.
* **Simple Redaction (RAR):** RAR is necessary but insufficient for auditability, as a redacted log doesn't provide cryptographic proof that no *other* unauthorized actions were taken.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Prover itself must run in a secure enclave (HAPE) to ensure it cannot be coerced into generating false proofs.
* **Observability:** Proof generation latency is monitored to ensure it doesn't cause "Cognitive Stall" in the agent.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
