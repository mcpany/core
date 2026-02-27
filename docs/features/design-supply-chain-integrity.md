# Design Doc: Skill Signature Verification (SSV) / Supply Chain Integrity Guard

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The rise of autonomous agent "Skills" (executable plugins) has introduced a major supply chain attack vector. Malicious MCP servers or skills can execute arbitrary code (RCE) or exfiltrate sensitive tokens when registered. OpenClaw and Claude Code have both seen recent exploits targeting these mechanisms. MCP Any needs a robust way to verify the provenance and integrity of every tool-providing entity before registration.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement cryptographic verification (SHA-256/ED25519) for all registered MCP server binaries and configuration files.
    *   Support "Attested Registration": tools are only exposed if their source matches a trusted signature.
    *   Provide an audit log of all registration attempts, including failed signature checks.
    *   Integrate with the Policy Engine to enforce "Signature-Only" execution modes.
*   **Non-Goals:**
    *   Providing a public PKI infrastructure (users provide their own trusted public keys).
    *   Sandboxing the execution (handled by the Command Adapter and OS-level policies).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Enterprise SysAdmin.
*   **Primary Goal:** Prevent an engineer from accidentally running a malicious MCP server cloned from an untrusted GitHub repo.
*   **The Happy Path (Tasks):**
    1.  Admin configures MCP Any with a list of "Trusted Public Keys."
    2.  Engineer attempts to add a new MCP server: `mcpany add --server my-tool --signature <sig>`.
    3.  MCP Any calculates the hash of the `my-tool` binary and config.
    4.  MCP Any verifies the signature against the hash using the trusted public keys.
    5.  Verification succeeds; the tools are registered and available to agents.
    6.  If an attacker modifies the binary later, MCP Any detects the hash mismatch during the next hot-reload and disables the tool.

## 4. Design & Architecture
*   **System Flow:**
    - **Manifest Generation**: Tools/Servers must include a `manifest.json` containing hashes of all executable components.
    - **Verification Loop**: The Service Registry invokes the `SSVMiddleware` during the `LoadService` lifecycle.
    - **Policy Enforcement**: If `STRICT_SIGNATURE_MODE=true`, any service without a valid signature is rejected.
*   **APIs / Interfaces:**
    - New field in service config: `signature: string`, `publicKeyId: string`.
    - Internal `IntegrityVerifier` interface in the Go backend.
*   **Data Storage/State:** Trusted keys and service hashes are stored in the secure `mcpany.db`.

## 5. Alternatives Considered
*   **Runtime Sandboxing (Docker/Wasm)**: *Rejected* as a primary solution because it's heavy and doesn't prevent "logical" exfiltration (e.g., sending tokens to a remote URL via valid network calls). Signature verification is a lighter, "shift-left" security control.
*   **Manual Approval (HITL)**: *Rejected* as the sole mechanism because it doesn't scale and is prone to human error/fatigue.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** SSV is a prerequisite for Zero-Trust discovery. Unverified tools are invisible to the search index.
*   **Observability:** Failed verification attempts trigger high-severity alerts in the dashboard.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation. Triggered by OpenClaw skill exploits and Claude Code CVE-2025-59536.
