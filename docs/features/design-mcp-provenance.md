# Design Doc: Supply Chain Integrity Guard (Provenance Attestation)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The "Clinejection" and other supply chain attacks have shown that MCP servers can be a vector for malicious tool injection. Since MCP Any allows for dynamic, configuration-based tool registration, it must ensure that every connected MCP server is legitimate and has not been tampered with. This document defines the "Supply Chain Integrity Guard," which implements cryptographic provenance attestation for all MCP tool sources.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Verify the cryptographic signature of MCP server configurations and binaries.
    *   Maintain a "Trust Registry" of known-good MCP server origins.
    *   Block tool execution from unverified or untrusted sources.
    *   Provide clear audit logs for attestation failures.
*   **Non-Goals:**
    *   Scanning the code of the MCP server for vulnerabilities (this is a provenance check, not a static analysis tool).
    *   Providing an "App Store" for MCP servers (MCP Any remains a gateway).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Operator.
*   **Primary Goal:** Prevent a rogue local script from impersonating a trusted MCP server and gaining access to sensitive tools.
*   **The Happy Path (Tasks):**
    1.  Operator adds a new MCP server configuration to MCP Any.
    2.  MCP Any identifies the `provenance_url` or `signature` field in the config.
    3.  Integrity Guard fetches the public key from the trusted provider (e.g., GitHub, Google, or a private PKI).
    4.  Integrity Guard verifies that the MCP server binary or config matches the signed manifest.
    5.  Tools are registered and marked as "Verified" in the UI.

## 4. Design & Architecture
*   **System Flow:**
    - **Pre-flight Check**: Before loading a service, the `AttestationMiddleware` triggers the `IntegrityGuard`.
    - **Validation**: The guard checks the service SHA256 against the `mcp.sig` file.
    - **Runtime Enforcement**: The `Policy Firewall` is updated to include a `is_verified` property for every tool.
*   **APIs / Interfaces:**
    - **Internal**: `IntegrityGuard.Verify(serviceConfig)`
    - **Configuration**: New `integrity` block in `config.yaml` to define trusted roots and enforcement levels (Warn vs. Block).
*   **Data Storage/State:** Verified service hashes are cached in the internal SQLite store to prevent repeated network calls for public keys.

## 5. Alternatives Considered
*   **Manual Whitelisting**: Forcing users to manually approve every tool. *Rejected* as too high-friction for large swarms.
*   **Sandboxing Only**: Relying purely on OS-level sandboxing (Docker/gVisor). *Rejected* because even a sandboxed tool can cause damage if it has access to authorized credentials.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core pillar of Zero Trust. Even if an agent is authorized to call a tool, the tool itself must be verified.
*   **Observability:** The UI must display a "Shield" icon next to verified tools. Attestation failures must trigger high-priority alerts.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
