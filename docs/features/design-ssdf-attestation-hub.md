# Design Doc: Automated SSDF Attestation Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of Autonomous PR Remediation (APR) in agents like Claude Code v3.3.0, the "Universal Agent Bus" must now bridge the gap between autonomous code generation and enterprise security compliance. Enterprise CI/CD pipelines are increasingly mandating adherence to the Secure Software Development Framework (SSDF).

The Automated SSDF Attestation Hub provides the infrastructure to generate and verify hardware-attested compliance fragments. These fragments serve as cryptographic proof that an autonomous code fix has been audited against SSDF standards and signed by a verified hardware authority before it reaches the codebase.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the generation of hardware-attested (TPM-signed) compliance fragments for AI-generated code.
    * Ensure autonomous PRs provide a verifiable "Compliance Provenance" trail.
    * Integrate with enterprise CI/CD gates to provide "Safe-to-Merge" signals based on SSDF criteria.
    * Support Recursive Provenance Compression (RPC) for compliance metadata in multi-hop swarms.
* **Non-Goals:**
    * Replacing existing Static/Dynamic Analysis (SAST/DAST) tools (it consumes their outputs).
    * Providing general-purpose code signing; it is specifically for "Agentic Compliance Provenance."

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise DevOps Security Engineer
* **Primary Goal:** Automatically verify and merge AI-generated security patches that comply with SSDF.
* **The Happy Path (Tasks):**
    1. A specialist agent (e.g., Claude Code) generates a security fix and submits it to MCP Any.
    2. The SSDF Attestation Hub intercepts the diff and triggers a "Compliance Audit" using verified local security tools.
    3. The Hub consumes the audit results and generates an SSDF-compliant attestation fragment.
    4. The fragment is cryptographically signed using the node's TPM and attached to the PR metadata.
    5. The enterprise CI/CD pipeline (e.g., GitHub Actions) calls the MCP Any verification endpoint.
    6. The Hub verifies the hardware signature and the compliance lineage.
    7. The PR is automatically merged once the SSDF "Safe-to-Merge" signal is confirmed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Fix] --> B[SSDF Hub]
        B --> C[Compliance Audit Engine]
        C --> D[TPM Signing Service]
        D --> E[Compliance Fragment]
        E --> F[CI/CD Verification Gate]
        F --> G[Merge Decision]
    ```
* **APIs / Interfaces:**
    * `ssdf.GenerateAttestation(diff, toolOutputs) -> FragmentID`: Generates a signed compliance fragment.
    * `ssdf.VerifyProvenance(fragmentID, signature) -> Status`: Validates the hardware signature and SSDF lineage.
    * `ssdf.GetCompliancePolicy() -> Policy`: Retrieves the current mesh-resident compliance rules.
* **Data Storage/State:**
    * **Compliance Shard Store:** Hardware-locked storage for active compliance fragments.
    * **Audit Baseline Registry:** Registry of approved security tools and their expected SSDF output schemas.

## 5. Alternatives Considered
* **Manual PR Review:** Rejected as it becomes a bottleneck for machine-speed remediation (the "Delegation Gap").
* **Unsigned Metadata Tags:** Rejected because they can be easily spoofed by compromised subagents or "Shadow PRs."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Compliance fragments are cryptographically bound to the mission root and the specific hardware enclave.
* **Observability:** Compliance success/failure rates are exported to the "Autonomous Compliance Dashboard" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
