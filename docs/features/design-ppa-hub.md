# Design Doc: Privacy-Preserving Audit (PPA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As enterprise adoption of autonomous agents scales, the conflict between "Absolute Auditability" and "Context Privacy" has become a primary blocker. Auditors need to verify that agents adhered to safety policies (e.g., "Did not exfiltrate PII"), but raw reasoning traces often contain proprietary source code or sensitive data.

The Privacy-Preserving Audit (PPA) Hub utilizes Zero-Knowledge Proofs (ZKP) to allow external observers to verify reasoning integrity and policy compliance without exposing raw context fragments.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested, Zero-Knowledge auditing of agent reasoning paths.
    * Generate verifiable proofs that reasoning followed specific mission-root constraints.
    * Support "Blind Auditing" where auditors only see the policy compliance result.
    * Neutralize the need for full trace exfiltration during security reviews.
* **Non-Goals:**
    * Replacing real-time monitoring (handled by CSAD/AEM).
    * Encrypting agent monologues for storage (handled by SRM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Auditor
* **Primary Goal:** Verify that a code-refactoring swarm did not access the `secrets/` directory during a 4-hour autonomous session.
* **The Happy Path (Tasks):**
    1. Auditor requests a "Compliance Proof" for `session_id: REF-202`.
    2. PPA Hub retrieves the hardware-attested reasoning trace metadata.
    3. PPA Hub generates a Zero-Knowledge proof that none of the retrieved context IDs map to the restricted `secrets/` path.
    4. The proof is signed by the local TPM and provided to the Auditor.
    5. Auditor verifies the proof against the global mission manifest and MCP Any security policy.
    6. Compliance is confirmed without the Auditor ever seeing the code refactored by the agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Reasoning Trace] --> B[PPA Circuit Generator]
        C[Safety Policy] --> B
        B --> D[ZK-SNARK Prover]
        D --> E[Hardware-Signed Proof]
        E --> F[External Auditor]
        F --> G{Verified?}
    ```
* **APIs / Interfaces:**
    * `ppa.GenerateComplianceProof(sessionID, policyID) -> ZKProof`: Generates a privacy-preserving proof.
    * `ppa.VerifyProof(zkProof, missionManifest) -> bool`: Lightweight verification for external auditors.
* **Data Storage/State:**
    * **Proof Registry:** Immutable ledger of issued compliance proofs and their hardware signatures.

## 5. Alternatives Considered
* **Differential Privacy (DP):** Rejected because it introduces noise that can mask subtle policy violations. ZKPs provide deterministic proof of compliance.
* **Redacted Logs:** Rejected because manual or heuristic redaction is prone to "Context Splicing" and information leakage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The prover circuit itself is hardware-attested to ensure it cannot be manipulated to generate false-positive proofs.
* **Observability:** Integrated with the "Zero-Knowledge Audit Viewer" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
