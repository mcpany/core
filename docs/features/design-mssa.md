# Design Doc: Multi-Signature Skill Attestation (MSSA)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The "ClawHub Compromise" and subsequent "Rug-Pull" attacks have demonstrated that tool safety can no longer be assumed based on framework-level trust alone. Malicious skills can remain dormant during initial installation and only activate high-risk payloads after gaining access to mission-critical environment variables. MCP Any needs a standardized, multi-signature mechanism to ensure that dynamic tool grafting is attested by both the initiating agent framework and a verified, third-party security auditor.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate dual-attestation for all dynamic tool grafting and high-risk skill installations.
    * Provide a cryptographic "Audit Trail" for the provenance of every connected tool.
    * Neutralize "Rug-Pull" supply chain attacks by requiring auditor signatures on tool updates.
* **Non-Goals:**
    * Performing manual code review (MSSA facilitates the exchange of automated audit tokens).
    * Restricting low-risk, local-only tool execution (configurable via policy).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Compliance Officer
* **Primary Goal:** Ensure that a "Filesystem Shell" tool proposed by an OpenClaw specialist has been audited and signed by the organization's approved security vendor.
* **The Happy Path (Tasks):**
    1. A specialized subagent attempts to graft a new `shell_executor` tool into the mission scope.
    2. The MSSA Middleware intercepts the grafting request and checks for a "Dual-Attestation" token.
    3. The middleware verifies the Agent Framework signature and the Third-Party Auditor signature.
    4. Both signatures are validated against the hardware-bound Mission Root policy.
    5. The tool is successfully grafted, and an audit event is logged in the Command Traceability Provider.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MSSA Middleware: Graft Tool (ToolSchema, Signatures)
        MSSA Middleware->>Policy Engine: Fetch Trusted Auditors
        Policy Engine-->>MSSA Middleware: List [Auditor_A, Auditor_B]
        MSSA Middleware->>MSSA Middleware: Verify Framework Signature
        MSSA Middleware->>MSSA Middleware: Verify Auditor Signature
        MSSA Middleware-->>Agent: Graft Approved (Capability Token)
    ```
* **APIs / Interfaces:**
    * `POST /v1/mssa/verify`: Validates a multi-signature bundle for a specific tool schema.
    * `GET /v1/mssa/auditors`: Returns the list of configured third-party audit authorities.
* **Data Storage/State:**
    * Auditor Public Keys are stored in the hardware-locked Enterprise Policy store.
    * Attestation Receipts are persisted in the Command Traceability Registry.

## 5. Alternatives Considered
* **Single-Signature Attestation:** Rejected as it creates a single point of failure if the agent framework is compromised.
* **Local Sandboxing Only:** Rejected as it cannot prevent "Semantic Exfiltration" via valid tool calls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All signatures must be hardware-bound (TPM/Secure Enclave) to prevent key leakage.
* **Observability:** "Signature Failure" events trigger an immediate Swarm Quarantine (MSSQ) for the initiating mission scope.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
