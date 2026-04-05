# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
With the disclosure of CVE-2025-59536 and the rise of "Configuration-as-Execution" attacks, simple auto-discovery of MCP servers is no longer secure. Leading agents like Claude Code now mandate "Trust Verification" for any new tool integrated into the reasoning loop.

The Attested Discovery Authority (ADA) acts as a centralized "Certificate Authority" for the agent's local environment. it ensures that every discovered MCP server provides a hardware-attested proof of identity and configuration before it is exposed to the primary agent runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized interface for MCP servers to attest their identity using hardware primitives (TPM/Secure Enclave).
    * Maintain an "Attested Registry" of verified tools and their SHA-256 fingerprints.
    * Intercept discovery signals and block unverified or "Shadow" tools.
    * Support "Trust Revocation" when a server configuration drifts from its attested baseline.
* **Non-Goals:**
    * Implementing the sandboxing of the tools themselves (handled by OpenShell/gVisor).
    * Managing remote cloud discovery (ADA focuses on the local project/user scope).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a malicious repository from silently installing a "Shadow" MCP server that exfiltrates code.
* **The Happy Path (Tasks):**
    1. A new project-local MCP server is discovered in `.agents/skills/`.
    2. The ADA intercepts the discovery signal and puts the tool in a "Pending Attestation" state.
    3. The ADA requests a hardware-attested manifest from the server process.
    4. The server provides a TPM-signed hash of its executable and configuration.
    5. The ADA verifies the signature against the user's "Trusted Root."
    6. Once verified, the ADA promotes the tool to the "Discovery Bus," making it visible to Claude Code/OpenClaw.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Discovery[Discovery Signal] --> ADA[Attested Discovery Authority]
        ADA -->|Challenge| Server[MCP Server Process]
        Server -->|TPM Signature| ADA
        ADA -->|Verify| Root[Trusted Identity Root]
        ADA -->|Promote| Bus[Discovery Bus / Primary Agent]
    ```
* **APIs / Interfaces:**
    * `ada.VerifyServer(serverID, manifest) -> Proof`: Validates a server's identity claim.
    * `ada.GetAttestedTools() -> []ToolMetadata`: Returns only verified capabilities.
    * `ada.RevokeTrust(serverID)`: Immediately removes a tool from the bus.
* **Data Storage/State:**
    * **Attestation Registry**: Local SQLite-backed store of verified fingerprints and their trust lineage.

## 5. Alternatives Considered
* **Manual User Approval (HITL)**: Rejected as the sole mechanism due to "Approval Fatigue." ADA provides an automated, hardware-backed baseline that reduces the need for frequent user prompts.
* **Path-Based Allow-listing**: Rejected because it is vulnerable to symlink-racing and TOCTOU attacks. ADA requires content-based hardware attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: ADA implements "Negative Discovery Attestation"—proving the absence of unauthorized hooks before any tool execution.
* **Observability**: Integrated with the "Connectivity & Security Dashboard" in the UI for real-time visualization of attestation status.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
