# Design Doc: Metadata Lineage Attestor
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
With the maturation of the Universal Agent Bus (UAB) and deep subagent swarms, "Metadata Context Poisoning" (CVE-2026-31201) has emerged as a high-severity threat. A compromised or malicious subagent can inject forged "Mission Metadata"—such as simulated system instructions or altered permission scopes—into the shared context handoff.

The Metadata Lineage Attestor ensures that every piece of structural or mission-critical metadata is cryptographically signed by its originator. This prevents "Parent Takeover" by ensuring the orchestrator can distinguish between authoritative system directives and agent-generated data.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement mandatory cryptographic signing for all "System-Level" mission metadata.
    *   Verify the provenance of metadata fragments during Binary State Handoff (BSH).
    *   Provide a "Redaction Layer" that blocks un-attested metadata from reaching the LLM's reasoning engine.
*   **Non-Goals:**
    *   Validating the truthfulness of natural language text within tool outputs (this is handled by IDS).
    *   Replacing the existing Policy Firewall (it complements it).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Prevent a specialized "Code Optimizer" subagent from granting itself "File:Write" permissions by injecting forged metadata into the parent mission's context.
*   **The Happy Path (Tasks):**
    1.  The parent agent initiates a "Read-Only" mission.
    2.  The Attestor signs the initial "Read-Only" scope metadata with the System Key.
    3.  A subagent attempts to hand back a state fragment containing a forged "Read-Write" scope.
    4.  The Attestor detects the signature mismatch during the BSH handoff.
    5.  The forged metadata is automatically redacted, and a security alert is triggered.
    6.  The parent agent receives the sanitized context and continues safely.

## 4. Design & Architecture
*   **System Flow:**
    `Metadata Source` -> `Signer` -> `Encapsulated Fragment (Data + Signature)` -> `BSH Gateway` -> `Attestor (Verify)` -> `LLM/Orchestrator`.
*   **APIs / Interfaces:**
    *   `mcp.metadata.sign(blob)`: Internal service for signing authoritative metadata.
    *   `mcp.metadata.verify(fragment)`: Middleware hook for the BSH pipeline.
*   **Data Storage/State:** Uses the server's master key (or TPM-bound key) for signing. Verification keys are rotated per session.

## 5. Alternatives Considered
*   **Schema-Only Validation:** Checking that metadata follows a specific JSON schema. Rejected because schemas don't prove *who* wrote the data.
*   **Isolation-Only:** Completely blocking subagents from seeing metadata. Rejected because subagents often need "Mission Context" to perform their tasks correctly.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Attestor is the primary defense against "Agentic Social Engineering" via metadata forgery.
*   **Observability:** All signature failures are logged with high severity in the `Security Audit Log`.

## 7. Evolutionary Changelog
*   **2026-04-11:** Initial Document Creation.
