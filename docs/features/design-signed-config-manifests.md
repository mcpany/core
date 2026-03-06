# Design Doc: Signed Configuration Manifests
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As AI agents (like Claude Code, OpenClaw) increasingly rely on project-level configuration files (e.g., `.mcp.json`, `.claude/settings.json`), these files have become a high-value target for attackers. Recent vulnerabilities (CVE-2025-59536, CVE-2026-21852) show that malicious configuration changes can lead to Remote Code Execution (RCE) and credential theft.

MCP Any must ensure that any configuration it ingests is authentic and has not been tampered with. This design introduces "Signed Configuration Manifests," requiring cryptographic proof of origin for configuration files before they are granted execution or elevated access privileges.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent RCE via malicious configuration hijacking.
    * Ensure integrity of project-level tool definitions.
    * Provide a clear "Untrusted" state for unsigned configurations.
    * Enable seamless developer experience via local signing keys.
* **Non-Goals:**
    * Implementing a full PKI (Public Key Infrastructure).
    * Signing every individual tool call (covered by Provenance Attestation).
    * Preventing all forms of prompt injection.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely collaborate on a repository containing MCP configurations without risking RCE from malicious PRs.
* **The Happy Path (Tasks):**
    1. Developer initializes a local signing key using `mcpany keys init`.
    2. Developer signs their `.mcp.yaml` config using `mcpany sign .mcp.yaml`.
    3. MCP Any server detects the signature manifest (`.mcp.yaml.sig`).
    4. Server validates the signature against the developer's public key (stored in a trusted local store or VCS-linked identity).
    5. Server loads the config with "Full Trust" status.

## 4. Design & Architecture
* **System Flow:**
    1. **Watcher Service**: Monitors filesystem for config changes.
    2. **Integrity Guard**: Checks for existence of `.sig` or `.asc` files corresponding to the config.
    3. **Trust Evaluator**:
        * If Valid Signature: Load with configured permissions.
        * If Missing/Invalid: Load in "Sandboxed Mode" (No shell execution, no environment variable access).
* **APIs / Interfaces:**
    * `ConfigLoader.Verify(path string) (TrustLevel, error)`
    * `CLI: mcpany sign [file]`
* **Data Storage/State:**
    * Trusted public keys stored in `$HOME/.mcpany/trusted_keys/`.
    * Trust state cached in memory per session.

## 5. Alternatives Considered
* **Manual Approval (HITL)**: Rejected as the primary mechanism due to friction, though it remains a fallback for unsigned configs.
* **Hardcoded Hash in Main Config**: Rejected as it is too brittle for active development.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Follows the principle of "Verify, then Trust." Unsigned configs are isolated in a restricted execution environment.
* **Observability:** Audit logs will record every signature verification attempt and trust escalation.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
