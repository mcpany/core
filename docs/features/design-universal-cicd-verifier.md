# Design Doc: Universal CI/CD Proof Verifier
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents increasingly participate in software development lifecycles, verifying the integrity of agent-generated code and pull requests has become a major security challenge. Enterprises need to ensure that an agent's reasoning followed strict security policies (e.g., no hardcoded secrets, no unauthorized library usage) before merging code into production.

The Universal CI/CD Proof Verifier provides a standardized way to ingest and verify Zero-Knowledge reasoning proofs (like Gemini's PPRP or MCP Any's PPA) within popular CI/CD platforms (GitHub Actions, GitLab CI).

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a universal interface for verifying ZK reasoning proofs from multiple agent frameworks.
    * Automate the "Go/No-Go" decision for agent-generated PRs based on proof validity.
    * Maintain developer privacy by verifying proofs without requiring access to the full agent context or codebase.
    * Integration with GitHub Actions and GitLab Runners.
* **Non-Goals:**
    * Generating the proofs themselves (handled by framework-specific providers).
    * Performing static analysis (SAST) on the code (handled by existing tools).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Compliance Officer
* **Primary Goal:** Automatically block any agent-generated PR that cannot prove it was generated under the "Secure Coding Standard" mission root.
* **The Happy Path (Tasks):**
    1. An agent generates a PR and attaches a PPA (Privacy-Preserving Audit) proof to the metadata.
    2. The CI/CD pipeline triggers the Universal Proof Verifier action.
    3. The Verifier extracts the proof and the associated mission-root hash.
    4. The Verifier calls the MCP Any PPA Hub to validate the cryptographic integrity of the proof.
    5. If valid, the CI check passes; otherwise, it fails and blocks the merge.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent PR] -->|Proof Metadata| B[CI/CD Runner]
        B --> C[Universal CI/CD Verifier]
        C --> D[MCP Any PPA Hub]
        D -->|Validation Result| C
        C -->|Check Pass/Fail| E[GitHub/GitLab Check]
    ```
* **APIs / Interfaces:**
    * `/api/v1/verify-proof`: Endpoint for the CI/CD runner to submit a proof for validation.
    * `verify-action@v1`: GitHub Action wrapper for the verification logic.
* **Data Storage/State:**
    * **Audit Log:** Immutable record of verification attempts and results.

## 5. Alternatives Considered
* **Manual Review of Reasoning Traces:** Rejected due to "Audit Fatigue" and privacy concerns (traces contain sensitive proprietary logic).
* **Framework-Specific Verifiers:** Rejected because enterprises use multiple frameworks (Claude Code, Gemini CLI, OpenClaw) and need a single source of truth for compliance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The verifier itself only sees the proof, not the source code, ensuring a Zero-Knowledge boundary.
* **Observability:** Verification results are pushed to the "CI/CD Compliance Auditor" UI component.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
